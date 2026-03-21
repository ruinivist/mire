#!/bin/sh
set -eu

# hoisted vars to fail fast if any missing
host_home=${MIRO_HOST_HOME:?}
host_tmp=${MIRO_HOST_TMP:?}
path_env=${MIRO_PATH_ENV:?}
visible_home=${MIRO_VISIBLE_HOME:?}
bootstrap_rc="$host_home/.miro-shell-rc"
visible_bootstrap_rc="$visible_home/.miro-shell-rc"
setup_scripts_dir='/tmp/miro-setup-scripts'

cat >"$bootstrap_rc" <<'EOF'
cd "${HOME:?}"

for path in /tmp/miro-setup-scripts/*.sh; do
  [ -e "$path" ] || continue
  cd "${HOME:?}"
  source "$path"
  cd "${HOME:?}"
done
EOF

if [ "${MIRO_COMPARE_MARKER:-0}" = "1" ]; then
  printf '__MIRO_E2E_BEGIN__\n'
fi

set -- \
  --ro-bind / / \
  --tmpfs /home \
  --bind "$host_home" "$visible_home" \
  --bind "$host_tmp" '/tmp' \
  --dev /dev \
  --proc /proc \
  --unshare-pid \
  --die-with-parent \
  --setenv HISTFILE '/dev/null' \
  --setenv HOME "$visible_home" \
  --setenv LANG 'C' \
  --setenv LC_ALL 'C' \
  --setenv PAGER 'cat' \
  --setenv PATH "$path_env" \
  --setenv PS1 '$ ' \
  --setenv TERM 'xterm-256color' \
  --setenv TMPDIR '/tmp' \
  --setenv TZ 'UTC' \
  --chdir "$visible_home"

if [ -n "${MIRO_SETUP_SCRIPTS:-}" ]; then
  i=1
  while IFS= read -r host_path || [ -n "$host_path" ]; do
    [ -n "$host_path" ] || continue
    visible_path=$(printf '%s/%03d.sh' "$setup_scripts_dir" "$i")
    set -- "$@" --ro-bind "$host_path" "$visible_path"
    i=$((i + 1))
  done <<EOF
${MIRO_SETUP_SCRIPTS-}
EOF
fi

exec bwrap "$@" bash --noprofile --rcfile "$visible_bootstrap_rc" -i
