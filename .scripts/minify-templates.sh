#!/bin/bash

# ============================================================================
# Roled Template Minification Script
# ============================================================================
# This script minifies CSS and JavaScript files used in Go templates while
# preserving Go template syntax ({{.Variable}}). It also creates minified
# versions of HTML templates.
#
# Usage:
#   ./minify-templates.sh [--css] [--js] [--html] [--all] [--check]
#
# Options:
#   --css    Minify CSS files only
#   --js     Minify JavaScript files only
#   --html   Minify HTML template files only
#   --all    Minify all files (default)
#   --check  Check if minification tools are installed
#
# Requirements:
#   - csso (CSS minifier): npm install -g csso-cli
#   - terser (JS minifier): npm install -g terser
#   - html-minifier (HTML minifier): npm install -g html-minifier-terser
#
# CI/CD Integration:
#   Add this script to your build pipeline before deployment:
#   - name: Minify templates
#     run: ./.scripts/minify-templates.sh --all
# ============================================================================

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
AUTH_DIR="${PROJECT_ROOT}/auth/internal/views"
ASSETS_DIR="${AUTH_DIR}/assets/static"
TEMPLATES_DIR="${AUTH_DIR}/templates"
ASSET_VERSION=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_tools() {
    local missing_tools=()
    
    if ! command -v csso &> /dev/null; then
        missing_tools+=("csso (npm install -g csso-cli)")
    fi
    
    if ! command -v terser &> /dev/null; then
        missing_tools+=("terser (npm install -g terser)")
    fi
    
    if ! command -v html-minifier-terser &> /dev/null; then
        missing_tools+=("html-minifier-terser (npm install -g html-minifier-terser)")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools:"
        for tool in "${missing_tools[@]}"; do
            echo "  - $tool"
        done
        echo ""
        echo "Install them with npm:"
        echo "  npm install -g csso-cli terser html-minifier-terser"
        exit 1
    fi
    
    log_success "All minification tools are installed"
    exit 0
}

# Protect Go template syntax by replacing it with placeholders
protect_go_templates() {
    local content="$1"
    # Replace Go template syntax with safe placeholders
    # Pattern: {{...}}
    content=$(echo "$content" | sed -E 's/\{\{/__double_curly_open__/g') # Replace {{ with __double_curly_open__
    content=$(echo "$content" | sed -E 's/\}\}/__double_curly_close__/g') # Replace }} with __double_curly_close__
    echo "$content"
}

# Restore Go template syntax from placeholders
restore_go_templates() {
    local content="$1"
    # Restore the Go template syntax
    content=$(echo "$content" | sed -E 's/__double_curly_open__/{{/g') # Replace __double_curly_open__ with {{
    content=$(echo "$content" | sed -E 's/__double_curly_close__/}}/g') # Replace __double_curly_close__ with }}
    echo "$content"
}

# Minify CSS files
minify_css() {
    log_info "Minifying CSS files..."
    
    local css_files=(
        "${ASSETS_DIR}/roled.css"
    )
    
    for css_file in "${css_files[@]}"; do
        if [ -f "$css_file" ]; then
            local filename=$(basename "$css_file")
            local dir=$(dirname "$css_file")
            local name="${filename%.css}"
            local output_file="${dir}/${name}.min.css"
            
            log_info "Processing: $filename"
            
            # CSS files don't typically contain Go templates, but we protect them anyway
            # for cases where CSS variables might use Go template values
            local content
            content=$(cat "$css_file")
            
            # Minify with csso
            echo "$content" | csso --output "$output_file"
            
            log_success "Created: ${name}.min.css"
        else
            log_warning "CSS file not found: $css_file"
        fi
    done
}

# Minify JavaScript files
minify_js() {
    log_info "Minifying JavaScript files..."
    
    local js_files=(
        "${ASSETS_DIR}/roled.js"
    )
    
    for js_file in "${js_files[@]}"; do
        if [ -f "$js_file" ]; then
            local filename=$(basename "$js_file")
            local dir=$(dirname "$js_file")
            local name="${filename%.js}"
            local output_file="${dir}/${name}.min.js"
            
            log_info "Processing: $filename"
            
            # Read file content
            local content
            content=$(cat "$js_file")
            
            # Protect Go template syntax
            local protected_content
            protected_content=$(protect_go_templates "$content")
            
            # Minify with terser
            echo "$protected_content" | terser --compress --mangle -o "$output_file"
            
            # Restore Go template syntax in the minified file
            if [ -f "$output_file" ]; then
                local minified_content
                minified_content=$(cat "$output_file")
                restored_content=$(restore_go_templates "$minified_content")
                echo "$restored_content" > "$output_file"
            fi
            
            log_success "Created: ${name}.min.js"
        else
            log_warning "JavaScript file not found: $js_file"
        fi
    done
}

# Minify HTML template files
minify_html() {
    log_info "Minifying HTML template files..."
    
    # Find all HTML templates (excluding already minified files)
    local html_files=()
    while IFS= read -r -d '' file; do
        # Skip files that are already minified (end with .min.html)
        if [[ ! "$file" =~ \.min\.html$ ]]; then
            html_files+=("$file")
        fi
    done < <(find "${TEMPLATES_DIR}" -name "*.html" -print0)
    
    for html_file in "${html_files[@]}"; do
        local filename=$(basename "$html_file")
        local dir=$(dirname "$html_file")
        local name="${filename%.html}"
        local output_file="${dir}/${name}.min.html"
        
        log_info "Processing: $filename"
        
        # Read file content
        local content
        content=$(cat "$html_file")
        
        # Protect Go template syntax
        local protected_content
        protected_content=$(protect_go_templates "$content")
        
        # Create temp file for minification
        local temp_file=$(mktemp)
        echo "$protected_content" > "$temp_file"
        
        # Minify with html-minifier-terser
        # Options:
        # --collapse-whitespace: Collapse whitespace
        # --remove-comments: Remove comments
        # --remove-optional-tags: Remove optional tags
        # --remove-redundant-attributes: Remove redundant attributes
        # --remove-script-type-attributes: Remove type="text/javascript"
        # --remove-style-link-type-attributes: Remove type="text/css"
        # --use-short-doctype: Use short doctype
        # --minify-css: Minify inline CSS
        # --minify-js: Minify inline JavaScript
        html-minifier-terser \
            --collapse-whitespace \
            --remove-comments \
            --remove-optional-tags \
            --remove-redundant-attributes \
            --remove-script-type-attributes \
            --remove-style-link-type-attributes \
            --use-short-doctype \
            --minify-css true \
            --minify-js true \
            --case-sensitive \
            --output "$output_file" \
            "$temp_file"
        
        # Restore Go template syntax in the minified file
        if [ -f "$output_file" ]; then
            local minified_content
            minified_content=$(cat "$output_file")
            local restored_content
            restored_content=$(restore_go_templates "$minified_content")
            
            # Replace CSS references with minified versions and update version
            if [ -n "$ASSET_VERSION" ]; then
                restored_content=$(echo "$restored_content" | sed -E "s|/assets/static/([a-zA-Z0-9_-]+)\.css\?v=1|/assets/static/\1.min.css?v=${ASSET_VERSION}|g")
            else
                restored_content=$(echo "$restored_content" | sed -E 's|/assets/static/([a-zA-Z0-9_-]+)\.css\?v=1|/assets/static/\1.min.css?v=1|g')
            fi
            
            # Replace JS references with minified versions and update version
            if [ -n "$ASSET_VERSION" ]; then
                restored_content=$(echo "$restored_content" | sed -E "s|/assets/static/([a-zA-Z0-9_-]+)\.js\?v=1|/assets/static/\1.min.js?v=${ASSET_VERSION}|g")
            else
                restored_content=$(echo "$restored_content" | sed -E 's|/assets/static/([a-zA-Z0-9_-]+)\.js\?v=1|/assets/static/\1.min.js?v=1|g')
            fi

            # Update partial definitions in minified files (e.g., {{define "partials/head"}} -> {{define "partials/head.min"}})
            restored_content=$(echo "$restored_content" | sed -E 's/\{\{([ -]*)define "([^"]+)"([ -]*)\}\}/{{\1define "\2.min"\3}}/g')

            # Update partial calls/includes in minified files (e.g., {{template "partials/head" .}} -> {{template "partials/head.min" .}})
            restored_content=$(echo "$restored_content" | sed -E 's/\{\{([ -]*)template "([^".]+)"([ ]*)/{{\1template "\2.min"\3/g')
            
            echo "$restored_content" > "$output_file"
        fi
        
        # Cleanup temp file
        rm -f "$temp_file"
        
        log_success "Created: ${name}.min.html"
    done
}

# Create versioned files for cache busting
create_versioned_files() {
    log_info "Creating versioned files for cache busting..."
    
    local timestamp=$(date +%Y%m%d%H%M%S)
    
    # CSS files
    for css_file in "${ASSETS_DIR}"/*.min.css; do
        if [ -f "$css_file" ]; then
            local filename=$(basename "$css_file")
            local name="${filename%.min.css}"
            local versioned_file="${ASSETS_DIR}/${name}.${timestamp}.min.css"
            cp "$css_file" "$versioned_file"
            log_success "Created versioned: ${name}.${timestamp}.min.css"
        fi
    done
    
    # JS files
    for js_file in "${ASSETS_DIR}"/*.min.js; do
        if [ -f "$js_file" ]; then
            local filename=$(basename "$js_file")
            local name="${filename%.min.js}"
            local versioned_file="${ASSETS_DIR}/${name}.${timestamp}.min.js"
            cp "$js_file" "$versioned_file"
            log_success "Created versioned: ${name}.${timestamp}.min.js"
        fi
    done
}

# Generate file sizes report
generate_report() {
    echo ""
    echo "=============================================="
    echo "Minification Report"
    echo "=============================================="
    
    # CSS files
    echo ""
    echo "CSS Files:"
    for css_file in "${ASSETS_DIR}"/*.css; do
        if [ -f "$css_file" ]; then
            local filename=$(basename "$css_file")
            local size=$(wc -c < "$css_file" | awk '{printf "%.2f KB", $1/1024}')
            printf "  %-40s %s\n" "$filename" "$size"
        fi
    done
    
    # JS files
    echo ""
    echo "JavaScript Files:"
    for js_file in "${ASSETS_DIR}"/*.js; do
        if [ -f "$js_file" ]; then
            local filename=$(basename "$js_file")
            local size=$(wc -c < "$js_file" | awk '{printf "%.2f KB", $1/1024}')
            printf "  %-40s %s\n" "$filename" "$size"
        fi
    done
    
    # HTML files
    echo ""
    echo "HTML Templates:"
    while IFS= read -r -d '' html_file; do
        if [ -f "$html_file" ]; then
            local rel_path="${html_file#${TEMPLATES_DIR}/}"
            local size=$(wc -c < "$html_file" | awk '{printf "%.2f KB", $1/1024}')
            printf "  %-40s %s\n" "$rel_path" "$size"
        fi
    done < <(find "${TEMPLATES_DIR}" -name "*.html" -print0 | sort -z)
    
    echo ""
    echo "=============================================="
}

# Main function
main() {
    local do_css=false
    local do_js=false
    local do_html=false
    local do_all=true
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --css)
                do_css=true
                do_all=false
                shift
                ;;
            --js)
                do_js=true
                do_all=false
                shift
                ;;
            --html)
                do_html=true
                do_all=false
                shift
                ;;
            --all)
                do_all=true
                shift
                ;;
            --version)
                ASSET_VERSION="$2"
                shift 2
                ;;
            --check)
                check_tools
                ;;
            --help|-h)
                echo "Usage: $0 [--css] [--js] [--html] [--all] [--version VERSION] [--check]"
                echo ""
                echo "Options:"
                echo "  --css           Minify CSS files only"
                echo "  --js            Minify JavaScript files only"
                echo "  --html          Minify HTML template files only"
                echo "  --all           Minify all files (default)"
                echo "  --version VER   Set asset version for cache busting (e.g., commit hash)"
                echo "  --check         Check if minification tools are installed"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Run '$0 --help' for usage information"
                exit 1
                ;;
        esac
    done
    
    echo ""
    echo "=============================================="
    echo "Roled Template Minification"
    echo "=============================================="
    echo ""
    
    if [ -n "$ASSET_VERSION" ]; then
        log_info "Asset version: ${ASSET_VERSION}"
    fi
    
    # Run minification
    if [ "$do_all" = true ] || [ "$do_css" = true ]; then
        minify_css
    fi
    
    if [ "$do_all" = true ] || [ "$do_js" = true ]; then
        minify_js
    fi
    
    if [ "$do_all" = true ] || [ "$do_html" = true ]; then
        minify_html
    fi
    
    # Generate report
    generate_report
    
    log_success "Minification complete!"
}

# Run main function
main "$@"


# ============================================================================
# CI/CD Integration Example
# ============================================================================
# Add the following step to your GitHub Actions workflow (e.g., build.yml)
# to minify templates before building Docker images or deploying:
#
#   # Add this job after lint/test jobs
#   minify-templates:
#     name: Minify Templates
#     runs-on: ubuntu-latest
#     steps:
#       - uses: actions/checkout@v7
#       
#       - name: Set up Node.js
#         uses: actions/setup-node@v7
#         with:
#           node-version: "24"
#       
#       - name: Install minification tools
#         run: |
#           npm install -g csso-cli terser html-minifier-terser
#       
#       - name: Run minification script
#         run: ./.scripts/minify-templates.sh --all
#       
#       - name: Upload minified artifacts
#         uses: actions/upload-artifact@v7
#         with:
#           name: minified-templates
#           path: |
#             auth/internal/views/assets/static/*.min.css
#             auth/internal/views/assets/static/*.min.js
#             auth/internal/views/templates/*.min.html
#
# Then in your deploy workflow, download the minified artifacts:
#
#   deploy:
#     needs: [auth-test, minify-templates]
#     steps:
#       - uses: actions/checkout@v7
#       
#       - name: Download minified templates
#         uses: actions/download-artifact@v8
#         with:
#           name: minified-templates
#           path: auth/internal/views/
#       
#       - name: Build and push Docker image
#         run: ...
# ============================================================================
