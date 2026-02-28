#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[0;37m'
GRAY='\033[0;90m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

LOCAL_POLL_INTERVAL=10
API_POLL_INTERVAL=30
CHECKS_POLL_INTERVAL=5

clear_screen() {
    printf '\033[H\033[2J\033[3J'
}

print_header() {
    :
}

print_not_git_repo() {
    clear_screen
    print_header
    echo -e "${GRAY}Not a git repository${RESET}"
    echo ""
    echo -e "${DIM}cd to a git repo to see PR status${RESET}"
}

get_branch_info() {
    local branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
    if [[ -z "$branch" ]]; then
        return 1
    fi
    echo "$branch"
}

get_upstream_info() {
    local branch=$1
    local upstream=$(git rev-parse --abbrev-ref "$branch@{upstream}" 2>/dev/null)
    if [[ -z "$upstream" ]]; then
        echo ""
        return
    fi

    local ahead=$(git rev-list --count "$upstream..$branch" 2>/dev/null || echo "0")
    local behind=$(git rev-list --count "$branch..$upstream" 2>/dev/null || echo "0")

    local result=""
    if [[ "$ahead" -gt 0 ]]; then
        result="${GREEN}>${ahead}${RESET}"
    fi
    if [[ "$behind" -gt 0 ]]; then
        [[ -n "$result" ]] && result="$result "
        result="${result}${RED}<${behind}${RESET}"
    fi
    if [[ -z "$result" ]]; then
        result="${GREEN}in sync${RESET}"
    fi
    echo -e "$result"
}

is_pushed() {
    local branch=$1
    git rev-parse --verify "origin/$branch" &>/dev/null
    return $?
}

get_unpushed_count() {
    local branch=$1
    local upstream="origin/$branch"

    if ! git rev-parse --verify "$upstream" &>/dev/null; then
        local default_branch=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@')
        [[ -z "$default_branch" ]] && default_branch="main"
        git rev-list --count "origin/$default_branch..$branch" 2>/dev/null || echo "?"
    else
        git rev-list --count "$upstream..$branch" 2>/dev/null || echo "0"
    fi
}

print_branch_info() {
    local branch=$1
    local upstream_info=$2
    local unpushed=$3
    local is_pushed=$4

    echo -e "${BOLD}Branch${RESET}"
    echo -e "   ${CYAN}$branch${RESET}"

    if [[ -n "$upstream_info" ]]; then
        echo -e "   $upstream_info"
    fi

    if [[ "$is_pushed" == "false" && "$unpushed" -gt 0 ]]; then
        echo -e "   ${YELLOW}$unpushed unpushed commit(s)${RESET}"
    fi
    echo ""
}

check_pr_exists() {
    gh pr view --json number &>/dev/null
    return $?
}

get_pr_info() {
    gh pr view --json number,title,state,url,reviewDecision,comments,mergeable,isDraft,headRefName 2>/dev/null
}

print_pr_info() {
    local pr_json=$1

    local number=$(echo "$pr_json" | jq -r '.number')
    local title=$(echo "$pr_json" | jq -r '.title')
    local state=$(echo "$pr_json" | jq -r '.state')
    local url=$(echo "$pr_json" | jq -r '.url')
    local review=$(echo "$pr_json" | jq -r '.reviewDecision // "PENDING"')
    local comments=$(echo "$pr_json" | jq -r '.comments | length')
    local mergeable=$(echo "$pr_json" | jq -r '.mergeable')
    local is_draft=$(echo "$pr_json" | jq -r '.isDraft')

    echo -e "${BOLD}PR #$number${RESET}"

    if [[ ${#title} -gt 35 ]]; then
        title="${title:0:32}..."
    fi
    echo -e "   ${WHITE}$title${RESET}"

    local state_icon state_color
    case "$state" in
        OPEN)
            if [[ "$is_draft" == "true" ]]; then
                state_icon="[draft]"
                state_color="$GRAY"
                state="DRAFT"
            else
                state_icon="[open]"
                state_color="$GREEN"
            fi
            ;;
        MERGED)
            state_icon="[merged]"
            state_color="$MAGENTA"
            ;;
        CLOSED)
            state_icon="[closed]"
            state_color="$RED"
            ;;
    esac
    echo -e "   ${state_icon} ${state_color}${state}${RESET}"

    local review_icon review_color
    case "$review" in
        APPROVED)
            review_icon="[ok]"
            review_color="$GREEN"
            ;;
        CHANGES_REQUESTED)
            review_icon="[changes]"
            review_color="$YELLOW"
            ;;
        *)
            review_icon="[wait]"
            review_color="$GRAY"
            review="PENDING"
            ;;
    esac
    echo -e "   ${review_icon} Review: ${review_color}${review}${RESET}"

    if [[ "$comments" -gt 0 ]]; then
        echo -e "   ${comments} comment(s)"
    fi

    case "$mergeable" in
        MERGEABLE)
            echo -e "   ${GREEN}Ready to merge${RESET}"
            ;;
        CONFLICTING)
            echo -e "   ${RED}Conflicts detected${RESET}"
            ;;
        UNKNOWN)
            echo -e "   ${GRAY}Checking mergeability...${RESET}"
            ;;
    esac

    echo -e "   ${DIM}${url}${RESET}"
    echo ""
}

print_no_pr() {
    echo -e "${BOLD}Pull Request${RESET}"
    echo -e "   ${GRAY}No PR yet${RESET}"
    echo -e "   ${DIM}Create one with: gh pr create${RESET}"
    echo ""
}

print_checks_header() {
    echo -e "${BOLD}Checks${RESET}"
}

get_checks_info() {
    gh pr checks --json name,state,bucket 2>/dev/null
}

print_checks() {
    local checks_json=$1

    if [[ -z "$checks_json" ]] || [[ "$checks_json" == "[]" ]]; then
        echo -e "   ${GRAY}No checks${RESET}"
        return
    fi

    echo "$checks_json" | jq -r '.[] | "\(.name)\t\(.state)\t\(.bucket)"' | while IFS=$'\t' read -r name state bucket; do
        local icon color status

        if [[ ${#name} -gt 20 ]]; then
            name="${name:0:17}..."
        fi

        case "$bucket" in
            pass)
                icon="[ok]"
                color="$GREEN"
                status="pass"
                ;;
            fail)
                icon="[x]"
                color="$RED"
                status="fail"
                ;;
            skipping)
                icon="[-]"
                color="$GRAY"
                status="skip"
                ;;
            pending)
                case "$state" in
                    IN_PROGRESS)
                        icon="[~]"
                        color="$YELLOW"
                        status="running"
                        ;;
                    QUEUED)
                        icon="[.]"
                        color="$GRAY"
                        status="queued"
                        ;;
                    *)
                        icon="[.]"
                        color="$GRAY"
                        status="pending"
                        ;;
                esac
                ;;
            *)
                icon="[?]"
                color="$GRAY"
                status="$state"
                ;;
        esac

        printf "   ${icon} ${color}%-20s %s${RESET}\n" "$name" "$status"
    done
}

main() {
    while true; do
        if ! git rev-parse --git-dir &>/dev/null; then
            print_not_git_repo
            sleep $LOCAL_POLL_INTERVAL
            continue
        fi

        local branch=$(get_branch_info)
        if [[ -z "$branch" ]]; then
            print_not_git_repo
            sleep $LOCAL_POLL_INTERVAL
            continue
        fi

        local upstream_info=$(get_upstream_info "$branch")
        local unpushed=$(get_unpushed_count "$branch")
        local pushed="false"
        if is_pushed "$branch"; then
            pushed="true"
        fi

        if [[ "$pushed" == "false" ]]; then
            clear_screen
            print_header
            print_branch_info "$branch" "$upstream_info" "$unpushed" "$pushed"
            echo -e "${DIM}Push branch to check for PR${RESET}"
            echo -e "${DIM}Refreshing in ${LOCAL_POLL_INTERVAL}s...${RESET}"
            sleep $LOCAL_POLL_INTERVAL
            continue
        fi

        local pr_json=$(get_pr_info)

        if [[ -z "$pr_json" ]]; then
            clear_screen
            print_header
            print_branch_info "$branch" "$upstream_info" "$unpushed" "$pushed"
            print_no_pr
            echo -e "${DIM}Refreshing in ${API_POLL_INTERVAL}s...${RESET}"
            sleep $API_POLL_INTERVAL
            continue
        fi

        local checks_json=$(get_checks_info)

        clear_screen
        print_header
        print_branch_info "$branch" "$upstream_info" "$unpushed" "$pushed"
        print_pr_info "$pr_json"
        print_checks_header
        print_checks "$checks_json"

        echo ""
        echo -e "${DIM}Refreshing in ${CHECKS_POLL_INTERVAL}s...${RESET}"
        sleep $CHECKS_POLL_INTERVAL
    done
}

main
