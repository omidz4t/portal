# bash completion for portal
_portal() {
	local cur prev
	COMPREPLY=()
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[COMP_CWORD-1]}"
	local cmds="serve init link list completion version help"
	local opts="-c --config -f --folder -a --account -h --help --version"
	case "$prev" in
		-c|--config|-f|--folder)
			COMPREPLY=( $(compgen -f -- "$cur") )
			return 0
			;;
		completion)
			COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
			return 0
			;;
	esac
	if [[ "$cur" == -* ]]; then
		COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
	else
		COMPREPLY=( $(compgen -W "$cmds" -- "$cur") )
	fi
}
complete -F _portal portal
