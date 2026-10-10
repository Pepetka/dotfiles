'builtin' 'local' '-a' 'p10k_config_opts'
[[ ! -o 'aliases' ]] || p10k_config_opts+=('aliases')
[[ ! -o 'sh_glob' ]] || p10k_config_opts+=('sh_glob')
[[ ! -o 'no_brace_expand' ]] || p10k_config_opts+=('no_brace_expand')
'builtin' 'setopt' 'no_aliases' 'no_sh_glob' 'brace_expand'
() {
  emulate -L zsh -o extended_glob
  unset -m '(POWERLEVEL9K_*|DEFAULT_USER)~POWERLEVEL9K_GITSTATUS_DIR'
  [[ $ZSH_VERSION == (5.<1->*|<6->.*) ]] || return
  typeset -g POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(
    os_icon
    dir
    vcs
    newline
    prompt_char
  )
  typeset -g POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=(
    status
    command_execution_time
    background_jobs
    direnv
    virtualenv
    nvm
    node_version
    kubecontext
    terraform
    aws
    gcloud
    docker_context
    context
    time
    newline
  )
  typeset -g POWERLEVEL9K_MODE=nerdfont-v3
  typeset -g POWERLEVEL9K_ICON_PADDING=moderate
  typeset -g POWERLEVEL9K_BACKGROUND=
  typeset -g POWERLEVEL9K_{LEFT,RIGHT}_{LEFT,RIGHT}_WHITESPACE=
  typeset -g POWERLEVEL9K_{LEFT,RIGHT}_SUBSEGMENT_SEPARATOR=' '
  typeset -g POWERLEVEL9K_{LEFT,RIGHT}_SEGMENT_SEPARATOR=
  typeset -g POWERLEVEL9K_ICON_BEFORE_CONTENT=true
  typeset -g POWERLEVEL9K_PROMPT_ADD_NEWLINE=true
  typeset -g POWERLEVEL9K_MULTILINE_FIRST_PROMPT_PREFIX='%F{#545c7e}╭─'
  typeset -g POWERLEVEL9K_MULTILINE_NEWLINE_PROMPT_PREFIX='%F{#545c7e}├─'
  typeset -g POWERLEVEL9K_MULTILINE_LAST_PROMPT_PREFIX='%F{#545c7e}╰─'
  typeset -g POWERLEVEL9K_MULTILINE_FIRST_PROMPT_SUFFIX=
  typeset -g POWERLEVEL9K_MULTILINE_NEWLINE_PROMPT_SUFFIX=
  typeset -g POWERLEVEL9K_MULTILINE_LAST_PROMPT_SUFFIX=
  typeset -g POWERLEVEL9K_LEFT_PROMPT_FIRST_SEGMENT_START_SYMBOL=' '
  typeset -g POWERLEVEL9K_RIGHT_PROMPT_LAST_SEGMENT_END_SYMBOL=
  typeset -g POWERLEVEL9K_SHOW_RULER=false
  typeset -g POWERLEVEL9K_RULER_CHAR='─'
  typeset -g POWERLEVEL9K_RULER_FOREGROUND="#545c7e"
  typeset -g POWERLEVEL9K_MULTILINE_FIRST_PROMPT_GAP_CHAR='·'
  if [[ $POWERLEVEL9K_MULTILINE_FIRST_PROMPT_GAP_CHAR != ' ' ]]; then
    typeset -g POWERLEVEL9K_MULTILINE_FIRST_PROMPT_GAP_FOREGROUND="#545c7e"
    typeset -g POWERLEVEL9K_LEFT_PROMPT_LAST_SEGMENT_END_SYMBOL=' '
    typeset -g POWERLEVEL9K_RIGHT_PROMPT_FIRST_SEGMENT_START_SYMBOL=' '
    typeset -g POWERLEVEL9K_EMPTY_LINE_LEFT_PROMPT_FIRST_SEGMENT_END_SYMBOL='%{%}'
    typeset -g POWERLEVEL9K_EMPTY_LINE_RIGHT_PROMPT_FIRST_SEGMENT_START_SYMBOL='%{%}'
  fi
  typeset -g POWERLEVEL9K_OS_ICON_FOREGROUND=
  typeset -g POWERLEVEL9K_PROMPT_CHAR_OK_{VIINS,VICMD,VIVIS,VIOWR}_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_PROMPT_CHAR_ERROR_{VIINS,VICMD,VIVIS,VIOWR}_FOREGROUND="#ff757f"
  typeset -g POWERLEVEL9K_PROMPT_CHAR_{OK,ERROR}_VIINS_CONTENT_EXPANSION='❯'
  typeset -g POWERLEVEL9K_PROMPT_CHAR_{OK,ERROR}_VICMD_CONTENT_EXPANSION='❮'
  typeset -g POWERLEVEL9K_PROMPT_CHAR_{OK,ERROR}_VIVIS_CONTENT_EXPANSION='V'
  typeset -g POWERLEVEL9K_PROMPT_CHAR_{OK,ERROR}_VIOWR_CONTENT_EXPANSION='▶'
  typeset -g POWERLEVEL9K_PROMPT_CHAR_OVERWRITE_STATE=true
  typeset -g POWERLEVEL9K_PROMPT_CHAR_LEFT_PROMPT_LAST_SEGMENT_END_SYMBOL=''
  typeset -g POWERLEVEL9K_PROMPT_CHAR_LEFT_PROMPT_FIRST_SEGMENT_START_SYMBOL=
  typeset -g POWERLEVEL9K_DIR_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_SHORTEN_STRATEGY=truncate_to_unique
  typeset -g POWERLEVEL9K_SHORTEN_DELIMITER=
  typeset -g POWERLEVEL9K_DIR_SHORTENED_FOREGROUND="#636da6"
  typeset -g POWERLEVEL9K_DIR_ANCHOR_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_DIR_ANCHOR_BOLD=true
  local anchor_files=(
    .bzr
    .citc
    .git
    .hg
    .node-version
    .python-version
    .go-version
    .ruby-version
    .lua-version
    .java-version
    .perl-version
    .php-version
    .tool-versions
    .mise.toml
    .shorten_folder_marker
    .svn
    .terraform
    CVS
    Cargo.toml
    composer.json
    go.mod
    package.json
    stack.yaml
  )
  typeset -g POWERLEVEL9K_SHORTEN_FOLDER_MARKER="(${(j:|:)anchor_files})"
  typeset -g POWERLEVEL9K_DIR_TRUNCATE_BEFORE_MARKER=false
  typeset -g POWERLEVEL9K_SHORTEN_DIR_LENGTH=1
  typeset -g POWERLEVEL9K_DIR_MAX_LENGTH=60
  typeset -g POWERLEVEL9K_DIR_MIN_COMMAND_COLUMNS=40
  typeset -g POWERLEVEL9K_DIR_MIN_COMMAND_COLUMNS_PCT=50
  typeset -g POWERLEVEL9K_DIR_HYPERLINK=true
  typeset -g POWERLEVEL9K_DIR_SHOW_WRITABLE=v3
  typeset -g POWERLEVEL9K_VCS_BRANCH_ICON='\uF126 '
  typeset -g POWERLEVEL9K_VCS_UNTRACKED_ICON='?'
  function my_git_formatter() {
    emulate -L zsh
    if [[ -n $P9K_CONTENT ]]; then
      typeset -g my_git_format=$P9K_CONTENT
      return
    fi
    if (($1)); then
      local meta='%f'
      local clean='%F{#c3e88d}'
      local modified='%F{#ffc777}'
      local untracked='%F{#82aaff}'
      local conflicted='%F{#ff757f}'
    else
      local meta='%F{#545c7e}'
      local clean='%F{#545c7e}'
      local modified='%F{#545c7e}'
      local untracked='%F{#545c7e}'
      local conflicted='%F{#545c7e}'
    fi
    local res
    if [[ -n $VCS_STATUS_LOCAL_BRANCH ]]; then
      local branch=${(V)VCS_STATUS_LOCAL_BRANCH}
      (($#branch > 32)) && branch[13,-13]="…"
      res+="${clean}${(g::)POWERLEVEL9K_VCS_BRANCH_ICON}${branch//\%/%%}"
    fi
    if [[ -n $VCS_STATUS_TAG &&
      -z $VCS_STATUS_LOCAL_BRANCH ]] \
      ; then
      local tag=${(V)VCS_STATUS_TAG}
      (($#tag > 32)) && tag[13,-13]="…"
      res+="${meta}#${clean}${tag//\%/%%}"
    fi
    [[ -z $VCS_STATUS_LOCAL_BRANCH && -z $VCS_STATUS_TAG ]] &&
      res+="${meta}@${clean}${VCS_STATUS_COMMIT[1,8]}"
    if [[ -n ${VCS_STATUS_REMOTE_BRANCH:#$VCS_STATUS_LOCAL_BRANCH} ]]; then
      res+="${meta}:${clean}${(V)VCS_STATUS_REMOTE_BRANCH//\%/%%}"
    fi
    if [[ $VCS_STATUS_COMMIT_SUMMARY == (|*[^[:alnum:]])(wip|WIP)(|[^[:alnum:]]*) ]]; then
      res+=" ${modified}wip"
    fi
    if ((VCS_STATUS_COMMITS_AHEAD || VCS_STATUS_COMMITS_BEHIND)); then
      ((VCS_STATUS_COMMITS_BEHIND)) && res+=" ${clean}⇣${VCS_STATUS_COMMITS_BEHIND}"
      ((VCS_STATUS_COMMITS_AHEAD && !VCS_STATUS_COMMITS_BEHIND)) && res+=" "
      ((VCS_STATUS_COMMITS_AHEAD)) && res+="${clean}⇡${VCS_STATUS_COMMITS_AHEAD}"
    elif [[ -n $VCS_STATUS_REMOTE_BRANCH ]]; then
    fi
    ((VCS_STATUS_PUSH_COMMITS_BEHIND)) && res+=" ${clean}⇠${VCS_STATUS_PUSH_COMMITS_BEHIND}"
    ((VCS_STATUS_PUSH_COMMITS_AHEAD && !VCS_STATUS_PUSH_COMMITS_BEHIND)) && res+=" "
    ((VCS_STATUS_PUSH_COMMITS_AHEAD)) && res+="${clean}⇢${VCS_STATUS_PUSH_COMMITS_AHEAD}"
    ((VCS_STATUS_STASHES)) && res+=" ${clean}*${VCS_STATUS_STASHES}"
    [[ -n $VCS_STATUS_ACTION ]] && res+=" ${conflicted}${VCS_STATUS_ACTION}"
    ((VCS_STATUS_NUM_CONFLICTED)) && res+=" ${conflicted}~${VCS_STATUS_NUM_CONFLICTED}"
    ((VCS_STATUS_NUM_STAGED)) && res+=" ${modified}+${VCS_STATUS_NUM_STAGED}"
    ((VCS_STATUS_NUM_UNSTAGED)) && res+=" ${modified}!${VCS_STATUS_NUM_UNSTAGED}"
    ((VCS_STATUS_NUM_UNTRACKED)) && res+=" ${untracked}${(g::)POWERLEVEL9K_VCS_UNTRACKED_ICON}${VCS_STATUS_NUM_UNTRACKED}"
    ((VCS_STATUS_HAS_UNSTAGED == -1)) && res+=" ${modified}─"
    typeset -g my_git_format=$res
  }
  functions -M my_git_formatter 2>/dev/null
  typeset -g POWERLEVEL9K_VCS_MAX_INDEX_SIZE_DIRTY=-1
  typeset -g POWERLEVEL9K_VCS_DISABLED_WORKDIR_PATTERN='~'
  typeset -g POWERLEVEL9K_VCS_DISABLE_GITSTATUS_FORMATTING=true
  typeset -g POWERLEVEL9K_VCS_CONTENT_EXPANSION='${$((my_git_formatter(1)))+${my_git_format}}'
  typeset -g POWERLEVEL9K_VCS_LOADING_CONTENT_EXPANSION='${$((my_git_formatter(0)))+${my_git_format}}'
  typeset -g POWERLEVEL9K_VCS_{STAGED,UNSTAGED,UNTRACKED,CONFLICTED,COMMITS_AHEAD,COMMITS_BEHIND}_MAX_NUM=-1
  typeset -g POWERLEVEL9K_VCS_VISUAL_IDENTIFIER_COLOR="#c3e88d"
  typeset -g POWERLEVEL9K_VCS_LOADING_VISUAL_IDENTIFIER_COLOR="#545c7e"
  typeset -g POWERLEVEL9K_VCS_BACKENDS=(git)
  typeset -g POWERLEVEL9K_VCS_CLEAN_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_VCS_UNTRACKED_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_VCS_MODIFIED_FOREGROUND="#ffc777"
  typeset -g POWERLEVEL9K_STATUS_EXTENDED_STATES=true
  typeset -g POWERLEVEL9K_STATUS_OK=false
  typeset -g POWERLEVEL9K_STATUS_OK_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_STATUS_OK_VISUAL_IDENTIFIER_EXPANSION='✔'
  typeset -g POWERLEVEL9K_STATUS_OK_PIPE=false
  typeset -g POWERLEVEL9K_STATUS_OK_PIPE_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_STATUS_OK_PIPE_VISUAL_IDENTIFIER_EXPANSION='✔'
  typeset -g POWERLEVEL9K_STATUS_ERROR=false
  typeset -g POWERLEVEL9K_STATUS_ERROR_FOREGROUND="#ff757f"
  typeset -g POWERLEVEL9K_STATUS_ERROR_VISUAL_IDENTIFIER_EXPANSION='✘'
  typeset -g POWERLEVEL9K_STATUS_ERROR_SIGNAL=true
  typeset -g POWERLEVEL9K_STATUS_ERROR_SIGNAL_FOREGROUND="#ff757f"
  typeset -g POWERLEVEL9K_STATUS_VERBOSE_SIGNAME=false
  typeset -g POWERLEVEL9K_STATUS_ERROR_SIGNAL_VISUAL_IDENTIFIER_EXPANSION='✘'
  typeset -g POWERLEVEL9K_STATUS_ERROR_PIPE=true
  typeset -g POWERLEVEL9K_STATUS_ERROR_PIPE_FOREGROUND="#ff757f"
  typeset -g POWERLEVEL9K_STATUS_ERROR_PIPE_VISUAL_IDENTIFIER_EXPANSION='✘'
  typeset -g POWERLEVEL9K_COMMAND_EXECUTION_TIME_THRESHOLD=2
  typeset -g POWERLEVEL9K_COMMAND_EXECUTION_TIME_PRECISION=1
  typeset -g POWERLEVEL9K_COMMAND_EXECUTION_TIME_FOREGROUND="#ff966c"
  typeset -g POWERLEVEL9K_COMMAND_EXECUTION_TIME_FORMAT='d h m s'
  typeset -g POWERLEVEL9K_BACKGROUND_JOBS_VERBOSE=false
  typeset -g POWERLEVEL9K_BACKGROUND_JOBS_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_DIRENV_FOREGROUND="#ffc777"
  typeset -g POWERLEVEL9K_DISK_USAGE_NORMAL_FOREGROUND="#86e1fc"
  typeset -g POWERLEVEL9K_DISK_USAGE_WARNING_FOREGROUND="#ffc777"
  typeset -g POWERLEVEL9K_DISK_USAGE_CRITICAL_FOREGROUND="#ff757f"
  typeset -g POWERLEVEL9K_DISK_USAGE_WARNING_LEVEL=90
  typeset -g POWERLEVEL9K_DISK_USAGE_CRITICAL_LEVEL=95
  typeset -g POWERLEVEL9K_DISK_USAGE_ONLY_WARNING=false
  typeset -g POWERLEVEL9K_RAM_FOREGROUND="#86e1fc"
  typeset -g POWERLEVEL9K_SWAP_FOREGROUND="#c099ff"
  typeset -g POWERLEVEL9K_LOAD_WHICH=5
  typeset -g POWERLEVEL9K_LOAD_NORMAL_FOREGROUND="#86e1fc"
  typeset -g POWERLEVEL9K_LOAD_WARNING_FOREGROUND="#ffc777"
  typeset -g POWERLEVEL9K_LOAD_CRITICAL_FOREGROUND="#ff757f"
  typeset -g POWERLEVEL9K_CPU_ARCH_FOREGROUND="#ff966c"
  typeset -g POWERLEVEL9K_CONTEXT_ROOT_FOREGROUND="#ffc777"
  typeset -g POWERLEVEL9K_CONTEXT_{REMOTE,REMOTE_SUDO}_FOREGROUND="#ffc777"
  typeset -g POWERLEVEL9K_CONTEXT_FOREGROUND="#ffc777"
  typeset -g POWERLEVEL9K_CONTEXT_ROOT_TEMPLATE='%B%n@%m'
  typeset -g POWERLEVEL9K_CONTEXT_{REMOTE,REMOTE_SUDO}_TEMPLATE='%n@%m'
  typeset -g POWERLEVEL9K_CONTEXT_TEMPLATE='%n@%m'
  typeset -g POWERLEVEL9K_CONTEXT_{DEFAULT,SUDO}_{CONTENT,VISUAL_IDENTIFIER}_EXPANSION=
  typeset -g POWERLEVEL9K_VIRTUALENV_FOREGROUND="#86e1fc"
  typeset -g POWERLEVEL9K_VIRTUALENV_SHOW_PYTHON_VERSION=false
  typeset -g POWERLEVEL9K_VIRTUALENV_SHOW_WITH_PYENV=false
  typeset -g POWERLEVEL9K_VIRTUALENV_{LEFT,RIGHT}_DELIMITER=
  typeset -g POWERLEVEL9K_NVM_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_NVM_PROMPT_ALWAYS_SHOW=false
  typeset -g POWERLEVEL9K_NVM_SHOW_SYSTEM=true
  typeset -g POWERLEVEL9K_NODE_VERSION_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_NODE_VERSION_PROJECT_ONLY=true
  typeset -g POWERLEVEL9K_GO_VERSION_FOREGROUND="#86e1fc"
  typeset -g POWERLEVEL9K_GO_VERSION_PROJECT_ONLY=true
  typeset -g POWERLEVEL9K_RUST_VERSION_FOREGROUND="#86e1fc"
  typeset -g POWERLEVEL9K_RUST_VERSION_PROJECT_ONLY=true
  typeset -g POWERLEVEL9K_DOTNET_VERSION_FOREGROUND="#c099ff"
  typeset -g POWERLEVEL9K_DOTNET_VERSION_PROJECT_ONLY=true
  typeset -g POWERLEVEL9K_PHP_VERSION_FOREGROUND="#c099ff"
  typeset -g POWERLEVEL9K_PHP_VERSION_PROJECT_ONLY=true
  typeset -g POWERLEVEL9K_LARAVEL_VERSION_FOREGROUND="#ff757f"
  typeset -g POWERLEVEL9K_JAVA_VERSION_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_JAVA_VERSION_PROJECT_ONLY=true
  typeset -g POWERLEVEL9K_JAVA_VERSION_FULL=false
  typeset -g POWERLEVEL9K_PACKAGE_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_KUBECONTEXT_SHOW_ON_COMMAND='kubectl|helm|kubens|kubectx|oc|istioctl|kogito|k9s|helmfile|flux|fluxctl|stern|kubeseal|skaffold|kubent|kubecolor|cmctl|sparkctl'
  typeset -g POWERLEVEL9K_KUBECONTEXT_CLASSES=(
    '*' DEFAULT)
  typeset -g POWERLEVEL9K_KUBECONTEXT_DEFAULT_FOREGROUND="#c099ff"
  typeset -g POWERLEVEL9K_KUBECONTEXT_DEFAULT_CONTENT_EXPANSION=
  POWERLEVEL9K_KUBECONTEXT_DEFAULT_CONTENT_EXPANSION+='${P9K_KUBECONTEXT_CLOUD_CLUSTER:-${P9K_KUBECONTEXT_NAME}}'
  POWERLEVEL9K_KUBECONTEXT_DEFAULT_CONTENT_EXPANSION+='${${:-/$P9K_KUBECONTEXT_NAMESPACE}:#/default}'
  typeset -g POWERLEVEL9K_TERRAFORM_SHOW_DEFAULT=false
  typeset -g POWERLEVEL9K_TERRAFORM_CLASSES=(
    '*' OTHER)
  typeset -g POWERLEVEL9K_TERRAFORM_OTHER_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_TERRAFORM_VERSION_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_AWS_SHOW_ON_COMMAND='aws|awless|cdk|terraform|pulumi|terragrunt'
  typeset -g POWERLEVEL9K_AWS_CLASSES=(
    '*' DEFAULT)
  typeset -g POWERLEVEL9K_AWS_DEFAULT_FOREGROUND="#ff966c"
  typeset -g POWERLEVEL9K_AWS_CONTENT_EXPANSION='${P9K_AWS_PROFILE//\%/%%}${P9K_AWS_REGION:+ ${P9K_AWS_REGION//\%/%%}}'
  typeset -g POWERLEVEL9K_GCLOUD_SHOW_ON_COMMAND='gcloud|gcs|gsutil'
  typeset -g POWERLEVEL9K_GCLOUD_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_GCLOUD_PARTIAL_CONTENT_EXPANSION='${P9K_GCLOUD_PROJECT_ID//\%/%%}'
  typeset -g POWERLEVEL9K_GCLOUD_COMPLETE_CONTENT_EXPANSION='${P9K_GCLOUD_PROJECT_NAME//\%/%%}'
  typeset -g POWERLEVEL9K_GCLOUD_REFRESH_PROJECT_NAME_SECONDS=60
  typeset -g POWERLEVEL9K_DOCKER_CONTEXT_SHOW_ON_COMMAND='docker|docker-compose|docker-compose-v2|podman'
  typeset -g POWERLEVEL9K_DOCKER_CONTEXT_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_PUBLIC_IP_FOREGROUND="#ff966c"
  typeset -g POWERLEVEL9K_VPN_IP_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_VPN_IP_CONTENT_EXPANSION=
  typeset -g POWERLEVEL9K_VPN_IP_INTERFACE='(gpd|wg|(.*tun)|tailscale)[0-9]*|(zt.*)'
  typeset -g POWERLEVEL9K_VPN_IP_SHOW_ALL=false
  typeset -g POWERLEVEL9K_IP_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_IP_CONTENT_EXPANSION='$P9K_IP_IP${P9K_IP_RX_RATE:+ %F{#c3e88d}⇣$P9K_IP_RX_RATE}${P9K_IP_TX_RATE:+ %F{#ff966c}⇡$P9K_IP_TX_RATE}'
  typeset -g POWERLEVEL9K_IP_INTERFACE='[ew].*'
  typeset -g POWERLEVEL9K_PROXY_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_BATTERY_LOW_THRESHOLD=20
  typeset -g POWERLEVEL9K_BATTERY_LOW_FOREGROUND="#ff757f"
  typeset -g POWERLEVEL9K_BATTERY_{CHARGING,CHARGED}_FOREGROUND="#c3e88d"
  typeset -g POWERLEVEL9K_BATTERY_DISCONNECTED_FOREGROUND="#ffc777"
  typeset -g POWERLEVEL9K_BATTERY_STAGES='\UF008E\UF007A\UF007B\UF007C\UF007D\UF007E\UF007F\UF0080\UF0081\UF0082\UF0079'
  typeset -g POWERLEVEL9K_BATTERY_VERBOSE=false
  typeset -g POWERLEVEL9K_WIFI_FOREGROUND="#82aaff"
  typeset -g POWERLEVEL9K_TIME_FOREGROUND="#86e1fc"
  typeset -g POWERLEVEL9K_TIME_FORMAT='%D{%H:%M:%S}'
  typeset -g POWERLEVEL9K_TIME_UPDATE_ON_COMMAND=false
  function prompt_example() {
    p10k segment -f 208 -i '⭐' -t 'hello, %n'
  }
  function instant_prompt_example() {
    prompt_example
  }
  typeset -g POWERLEVEL9K_TRANSIENT_PROMPT=always
  typeset -g POWERLEVEL9K_INSTANT_PROMPT=quiet
  typeset -g POWERLEVEL9K_DISABLE_HOT_RELOAD=true
  ((!$+functions[p10k])) || p10k reload
}
typeset -g POWERLEVEL9K_CONFIG_FILE=${${(%):-%x}:a}
((${#p10k_config_opts})) && setopt ${p10k_config_opts[@]}
'builtin' 'unset' 'p10k_config_opts'
