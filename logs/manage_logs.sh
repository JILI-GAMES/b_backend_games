#!/bin/bash

# Log management script for per-game logging system
# This script helps manage log files, viewing, and cleanup
# Log files are created directly in the root directory (like app.log)

# Change to the project root directory to run from anywhere
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

LOG_DIR="logs"

GAME_LOGS=(
    "birdsparty.log"
    "birdspartydeluxe.log" 
    "blossomsofwealth.log"
    "crazykingkong.log"
    "funkykingkong.log"
    "hilo.log"
    "kong.log"
    "magicace.log"
    "magicaceoriginal.log"
    "moneybagsman.log"
    "moneybagsman2.log"
    "onepiece.log"
    "opensesame1.log"
    "opensesame2.log"
    "superace-deluxe.log"
    "winningmask.log"
    "server.log"
)

show_usage() {
    echo "📋 Log Management Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  list                    - List all log files and sizes"
    echo "  view [game]             - View recent logs for a specific game"
    echo "  tail [game]             - Tail logs for a specific game"
    echo "  search [game] [term]    - Search for term in game logs"
    echo "  errors [game]           - Show error logs for a specific game"
    echo "  cleanup                 - Clean up old log files"
    echo "  stats                   - Show logging statistics"
    echo ""
    echo "Examples:"
    echo "  $0 list"
    echo "  $0 view blossomsofwealth"
    echo "  $0 tail hilo"
    echo "  $0 search moneybagsman2 \"Failed to get RTP\""
    echo "  $0 errors onepiece"
}

list_logs() {
    echo "📁 Log Files (in logs/ directory):"
    echo "=================================="
    for log in "${GAME_LOGS[@]}"; do
        log_path="$LOG_DIR/$log"
        if [ -f "$log_path" ]; then
            size=$(du -h "$log_path" | cut -f1)
            lines=$(wc -l < "$log_path")
            echo "📄 $log - Size: $size, Lines: $lines"
        else
            echo "❌ $log - Not found"
        fi
    done
}

view_logs() {
    local game=$1
    if [ -z "$game" ]; then
        echo "❌ Please specify a game name"
        return 1
    fi
    
    local log_file="$LOG_DIR/${game}.log"
    if [ -f "$log_file" ]; then
        echo "📄 Recent logs for $game:"
        echo "========================="
        tail -50 "$log_file"
    else
        echo "❌ Log file not found: $log_file"
    fi
}

tail_logs() {
    local game=$1
    if [ -z "$game" ]; then
        echo "❌ Please specify a game name"
        return 1
    fi
    
    local log_file="$LOG_DIR/${game}.log"
    if [ -f "$log_file" ]; then
        echo "📄 Tailing logs for $game (Ctrl+C to stop):"
        echo "============================================"
        tail -f "$log_file"
    else
        echo "❌ Log file not found: $log_file"
    fi
}

search_logs() {
    local game=$1
    local term=$2
    
    if [ -z "$game" ] || [ -z "$term" ]; then
        echo "❌ Please specify game name and search term"
        return 1
    fi
    
    local log_file="$LOG_DIR/${game}.log"
    if [ -f "$log_file" ]; then
        echo "🔍 Searching for '$term' in $game logs:"
        echo "======================================="
        grep -i "$term" "$log_file" | tail -20
    else
        echo "❌ Log file not found: $log_file"
    fi
}

show_errors() {
    local game=$1
    if [ -z "$game" ]; then
        echo "❌ Please specify a game name"
        return 1
    fi
    
    local log_file="$LOG_DIR/${game}.log"
    if [ -f "$log_file" ]; then
        echo "🚨 Error logs for $game:"
        echo "========================"
        grep -i "error\|failed\|exception" "$log_file" | tail -20
    else
        echo "❌ Log file not found: $log_file"
    fi
}

cleanup_logs() {
    echo "🧹 Cleaning up old log files..."
    find "$LOG_DIR" -name "*.log.*" -mtime +7 -delete
    echo "✅ Cleanup complete"
}

show_stats() {
    echo "📊 Logging Statistics:"
    echo "====================="
    
    total_files=0
    total_size=0
    
    for log in "${GAME_LOGS[@]}"; do
        if [ -f "$LOG_DIR/$log" ]; then
            size=$(stat -f%z "$LOG_DIR/$log" 2>/dev/null || stat -c%s "$LOG_DIR/$log" 2>/dev/null)
            total_size=$((total_size + size))
            total_files=$((total_files + 1))
        fi
    done
    
    echo "📄 Total log files: $total_files"
    echo "💾 Total size: $(numfmt --to=iec $total_size)"
    echo "📅 Log directory: $LOG_DIR"
}

# Main script logic
case "$1" in
    "list")
        list_logs
        ;;
    "view")
        view_logs "$2"
        ;;
    "tail")
        tail_logs "$2"
        ;;
    "search")
        search_logs "$2" "$3"
        ;;
    "errors")
        show_errors "$2"
        ;;
    "cleanup")
        cleanup_logs
        ;;
    "stats")
        show_stats
        ;;
    *)
        show_usage
        ;;
esac
