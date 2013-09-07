# Example functions that can be used with zsh

function cd_log()
{
  cwd_logger
}

chpwd_functions=( "${chpwd_functions[@]}" cd_log )

function f () {
  if [[ $# -eq 1 ]]; then
    cd "$(cwd_frequency $1)"
  else
    cwd_frequency
  fi
}

function r () {
  if [[ $# -eq 1 ]]; then
    cd "$(cwd_recently $1)"
  else
    cwd_recently
  fi
}
