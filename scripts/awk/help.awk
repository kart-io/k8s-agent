# This script is called by `make help` to generate the help output.
# It parses all makefiles and prints a formatted list of targets.

BEGIN {
    # Global constants
    SPECIAL_PREFIXES = "_install\\.|_uninstall\\.|_verify\\.";

    # Color codes for consistent formatting
    COLOR_CYAN = "\033[36m";
    COLOR_MAGENTA = "\033[35m";
    COLOR_BOLD = "\033[1m";
    COLOR_RESET = "\033[0m";

    FS = ":[ \t]*##";  # Split on colon + spaces/tabs + ##
    current_category = "";

    printf "\nUsage:\n  make " COLOR_CYAN "<TARGETS> <OPTIONS>" COLOR_RESET "\n";
    printf "\n" COLOR_MAGENTA "Targets:" COLOR_RESET "\n";
}

# Function to extract file prefix from .mk files
function get_file_prefix(filename,    path_parts, result) {
    if (filename ~ /\.mk$/) {
        split(filename, path_parts, "/");
        result = path_parts[length(path_parts)];
        sub(/\.mk$/, "", result);
        return result;
    }
    return "";
}

# Function to apply file prefix rules and formatting
function format_target(name, file_prefix) {
    if (file_prefix != "") {
        # Replace special prefixes in one go
        if (name ~ "^(" SPECIAL_PREFIXES ")") {
            sub(/^_/, file_prefix ".", name);
        }
        gsub("_", ".", name); # Convert underscores to dots for .mk files
    } else {
        gsub("_", "-", name); # Convert underscores to hyphens for main Makefile
    }
    return name;
}

# Function to process category headers
function process_category_header(line, indent_prefix,    category_name) {
    category_name = substr(line, 5); # Remove "##@ " prefix
    printf "%s" COLOR_BOLD "%s" COLOR_RESET "\n", indent_prefix, category_name;
    return tolower(category_name);
}

# Function to extract target name from variable form $(VAR)
function extract_var_target(target,    var) {
    if (target ~ /^\$\(/) {
        var = substr(target, 3, length(target)-3);
        return var;
    }
    return "";
}

# Function to extract simple target name
function extract_simple_target(target,    arr) {
    split(target, arr, ":");
    return arr[1];
}

# Function to format target output line
function format_target_line(target_name, comment, indent_spaces) {
    printf "%s" COLOR_CYAN "%-45s" COLOR_RESET "%s\n",
           indent_spaces, target_name, comment;
}

# Process category headers
/^##@/ {
    current_category = process_category_header($0, "\n");
    next;
}

# Process documented targets
/^[^	 #].*:[ 	]*##/ {
    target = $1;
    comment = $2;
    file_prefix = get_file_prefix(FILENAME);

    # Variable form: $(VAR)
    var = extract_var_target(target);
    if (var != "") {
        var = format_target(var, file_prefix);
        format_target_line(tolower(var), comment, "  ");
    }
    # Simple target
    else {
        target_name = extract_simple_target(target);
        target_name = format_target(target_name, file_prefix);
        format_target_line(target_name, comment, "  ");
    }
}

# Print footer
END {
    if (ENVIRON["USAGE_OPTIONS"]) {
        printf "%s\n", ENVIRON["USAGE_OPTIONS"];
    }
}
