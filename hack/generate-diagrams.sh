#!/usr/bin/env bash
# hack/generate-diagrams.sh — Parses @doc annotations from Go source and
# generates Mermaid diagrams into docs/DIAGRAMS.md.
#
# Annotation syntax (in Go doc comments):
#
#   @relates  TypeA ||--o{ TypeB : "label"
#   @state    StateA --> StateB : trigger
#   @flow     StepA["Label"] --> StepB["Label"]
#   @component NodeID["Label"] --> TargetID["Label"]
#
# Output structure:
#   - Component Overview    (one sub-diagram per connected subgraph)
#   - Entity Relationships  (one sub-diagram per connected subgraph)
#   - State Transitions     (one sub-diagram per source file)
#   - Processing Flows      (one sub-diagram per source file)

set -euo pipefail

# Ensure deterministic sort order across macOS / Linux CI
export LC_ALL=C

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

output="${repo_root}/docs/DIAGRAMS.md"
tmp_file="$(mktemp)"
trap 'rm -f "${tmp_file}"' EXIT

# --- Helper: derive a human-readable heading from a Go source file path ---
file_to_heading() {
  local relpath="${1#${repo_root}/}"
  local base
  base="$(basename "$relpath" .go)"
  base="${base%_types}"
  base="${base%_controller}"
  base="${base%_storage}"
  echo "${base}" | sed 's/_/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1'
}

# --- Helper: extract package context from file path for heading ---
file_to_context() {
  local relpath="${1#${repo_root}/}"
  local dir
  dir="$(dirname "$relpath")"
  case "$dir" in
    */v1alpha1)   echo "CRD" ;;
    */controller) echo "Controller" ;;
    */agent)      echo "Agent" ;;
    */security)   echo "Security" ;;
    */usbip)      echo "USB/IP" ;;
    */backup)     echo "Backup" ;;
    */webhook)    echo "Webhook" ;;
    *)            echo "${dir##*/}" ;;
  esac
}

# --- Collect raw annotations with file path preserved ---
# Format: filepath<TAB>annotation_content
raw_relations="$(mktemp)"
raw_states="$(mktemp)"
raw_flows="$(mktemp)"
raw_components="$(mktemp)"
trap 'rm -f "${tmp_file}" "${raw_relations}" "${raw_states}" "${raw_flows}" "${raw_components}"' EXIT

grep -rn '@relates' --include='*.go' "${repo_root}" | grep -v '_test.go' | \
  sed "s|^\(${repo_root}/[^:]*\):.*@relates[[:space:]]*|\1	|" >> "${raw_relations}" || true

grep -rn '@state' --include='*.go' "${repo_root}" | grep -v '_test.go' | \
  sed "s|^\(${repo_root}/[^:]*\):.*@state[[:space:]]*|\1	|" >> "${raw_states}" || true

grep -rn '@flow' --include='*.go' "${repo_root}" | grep -v '_test.go' | \
  sed "s|^\(${repo_root}/[^:]*\):.*@flow[[:space:]]*|\1	|" >> "${raw_flows}" || true

grep -rn '@component' --include='*.go' "${repo_root}" | grep -v '_test.go' | \
  sed "s|^\(${repo_root}/[^:]*\):.*@component[[:space:]]*|\1	|" >> "${raw_components}" || true

# --- emit_clustered: split edges into connected subgraphs ---
# Uses union-find in awk to detect disconnected clusters
# and emits one Mermaid block per cluster with a derived heading.
#
# $1 = file with annotation lines (one per line, no file-path prefix)
# $2 = mermaid diagram header (e.g. "flowchart TB", "erDiagram")
# $3 = edge format: "flowchart" or "er"
emit_clustered() {
  local lines_file="$1" mermaid_header="$2" edge_format="$3"

  sort -u "${lines_file}" | awk -v header="${mermaid_header}" -v fmt="${edge_format}" '
    # --- Union-Find with path compression ---
    function find(x,  r, t) {
        r = x; while (par[r] != r) r = par[r]
        while (par[x] != r) { t = par[x]; par[x] = r; x = t }
        return r
    }
    function unite(a, b) { a = find(a); b = find(b); if (a != b) par[a] = b }

    # --- Extract first identifier from a string ---
    function first_word(s,  i) {
        gsub(/^[[:space:]]+/, "", s)
        i = match(s, /[^A-Za-z0-9_]/)
        return (i > 0) ? substr(s, 1, i - 1) : s
    }

    # --- Extract Mermaid label ["..."] or fall back to ID ---
    function extract_label(s,  p, lbl) {
        p = index(s, "[\"")
        if (p > 0) { lbl = substr(s, p + 2); sub(/"[\]].*/, "", lbl); return lbl }
        return first_word(s)
    }

    # --- Parse each edge line ---
    {
        N++; lines[N] = $0
        id_l = ""; id_r = ""

        if (fmt == "flowchart") {
            p = index($0, "-->")
            if (p > 0) {
                lhs = substr($0, 1, p - 1); rhs = substr($0, p + 3)
                id_l = first_word(lhs); id_r = first_word(rhs)
                lbl_l = extract_label(lhs); lbl_r = extract_label(rhs)
            }
        } else {
            # ER format: TypeA <operator> TypeB : "label"
            tmp = $0; gsub(/^[[:space:]]+/, "", tmp)
            id_l = first_word(tmp)
            # Remove label portion after " : "
            c = index(tmp, " : "); if (c > 0) tmp = substr(tmp, 1, c - 1)
            # Last word in the remaining string is the right type
            gsub(/[[:space:]]+$/, "", tmp)
            i = length(tmp)
            while (i > 0 && substr(tmp, i, 1) ~ /[A-Za-z0-9_]/) i--
            id_r = substr(tmp, i + 1)
            lbl_l = id_l; lbl_r = id_r
        }

        if (id_l != "" && id_r != "") {
            if (!(id_l in par)) par[id_l] = id_l
            if (!(id_r in par)) par[id_r] = id_r
            unite(id_l, id_r)
            el[N] = id_l; er[N] = id_r
            # Prefer descriptive labels over bare IDs
            if (!(id_l in lbl) || lbl[id_l] == id_l) lbl[id_l] = lbl_l
            if (!(id_r in lbl) || lbl[id_r] == id_r) lbl[id_r] = lbl_r
            deg[id_l]++; deg[id_r]++
        }
    }

    END {
        # --- Assign edges to groups ---
        gc = 0
        for (i = 1; i <= N; i++) {
            root = (el[i] != "") ? find(el[i]) : ("_s" i)
            if (!(root in gi)) { gc++; gi[root] = gc; gr[gc] = root }
            g = gi[root]; gcnt[g]++; ge[g, gcnt[g]] = i
        }

        # --- Name each group by its most-connected node label ---
        for (g = 1; g <= gc; g++) {
            root = gr[g]; best = ""; bestd = 0
            for (x in par) {
                if (find(x) == root && deg[x] > bestd) { bestd = deg[x]; best = lbl[x] }
            }
            gn[g] = (best != "") ? best : root
        }

        # --- Sort groups alphabetically by heading ---
        for (i = 1; i <= gc; i++) ord[i] = i
        for (i = 1; i < gc; i++)
            for (j = i + 1; j <= gc; j++)
                if (gn[ord[i]] > gn[ord[j]]) { t = ord[i]; ord[i] = ord[j]; ord[j] = t }

        # --- Output one sub-diagram per cluster ---
        for (oi = 1; oi <= gc; oi++) {
            g = ord[oi]
            print "### " gn[g]
            print ""
            print "```mermaid"
            print header
            for (k = 1; k <= gcnt[g]; k++) print "    " lines[ge[g, k]]
            print "```"
            print ""
        }
    }
  '
}

# --- emit_grouped: one sub-diagram per source file ---
# Used for @state and @flow annotations where per-file context matters.
emit_grouped() {
  local raw_file="$1" diagram_type="$2" mermaid_header="$3"
  local prev_file="" heading ctx relpath

  while IFS=$'\t' read -r fpath content; do
    if [ "${fpath}" != "${prev_file}" ]; then
      if [ -n "${prev_file}" ]; then
        echo '```'
        echo
      fi
      heading="$(file_to_heading "${fpath}")"
      ctx="$(file_to_context "${fpath}")"
      relpath="${fpath#${repo_root}/}"
      echo "### ${heading} (${ctx})"
      echo
      echo "_Source: \`${relpath}\`_"
      echo
      echo '```mermaid'
      echo "${mermaid_header}"
      prev_file="${fpath}"
    fi
    echo "    ${content}"
  done < <(sort -t$'\t' -k1,1 "${raw_file}")
  if [ -n "${prev_file}" ]; then
    echo '```'
    echo
  fi
}

# ==========================================================
# Output
# ==========================================================
{
  echo "# Auto-Generated Diagrams"
  echo
  echo "_Generated from \`@doc\` annotations in Go source by \`hack/generate-diagrams.sh\`._"
  echo
  echo "---"
  echo

  # --- Component Overview (clustered by connected subgraph) ---
  if [ -s "${raw_components}" ]; then
    echo "## Component Overview"
    echo
    tmp_comp="$(mktemp)"
    cut -f2 "${raw_components}" > "${tmp_comp}"
    emit_clustered "${tmp_comp}" "flowchart TB" "flowchart"
    rm -f "${tmp_comp}"
  fi

  # --- Entity Relationships (clustered by connected subgraph) ---
  if [ -s "${raw_relations}" ]; then
    echo "---"
    echo
    echo "## Entity Relationships"
    echo
    tmp_rel="$(mktemp)"
    cut -f2 "${raw_relations}" > "${tmp_rel}"
    emit_clustered "${tmp_rel}" "erDiagram" "er"
    rm -f "${tmp_rel}"
  fi

  # --- State Transitions (grouped per source file) ---
  if [ -s "${raw_states}" ]; then
    echo "---"
    echo
    echo "## State Transitions"
    echo
    emit_grouped "${raw_states}" "state" "stateDiagram-v2"
  fi

  # --- Processing Flows (grouped per source file) ---
  if [ -s "${raw_flows}" ]; then
    echo "---"
    echo
    echo "## Processing Flows"
    echo
    emit_grouped "${raw_flows}" "flow" "flowchart TD"
  fi

  # If no annotations found
  if [ ! -s "${raw_relations}" ] && [ ! -s "${raw_states}" ] && [ ! -s "${raw_flows}" ] && [ ! -s "${raw_components}" ]; then
    echo "> No \`@doc\` annotations found yet. Add annotations to Go doc comments to auto-generate diagrams."
    echo
  fi

} > "${tmp_file}"

mv "${tmp_file}" "${output}"
echo "Generated ${output}"
