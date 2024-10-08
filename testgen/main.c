#define _POSIX_C_SOURCE 200809L

#include <errno.h>
#include <getopt.h>
#include <jansson.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define ANSI_COLOR_RED "\x1b[31m"
#define ANSI_COLOR_RESET "\x1b[0m"

const char* INFILE_PATH = "testdata/cases.json";
const char* OUTFILE_PATH = "testdata/fixtures.json";

static void log_err(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    fprintf(stderr, ANSI_COLOR_RED);
    fprintf(stderr, "error: ");
    vfprintf(stderr, fmt, args);
    fprintf(stderr, ANSI_COLOR_RESET);
    fprintf(stderr, "\n");
    va_end(args);
}

static void log_info(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    vfprintf(stderr, fmt, args);
    fprintf(stderr, "\n");
    va_end(args);
}

static bool validate_label(const json_t* item, json_t** label, const size_t index) {
    *label = json_object_get(item, "label");
    if (!*label) {
        log_err("missing required prop [%ld].label", index);
        return true;
    }
    if (!json_is_string(*label)) {
        log_err("expected prop [%ld].label to be a json string", index);
        return true;
    }
    return false;
}

static bool validate_args(const json_t* item, json_t** args, const size_t index) {
    *args = json_object_get(item, "args");
    if (!*args) {
        log_err("missing required prop [%ld].args", index);
        return true;
    }
    if (!json_is_array(*args)) {
        log_err("expected prop [%ld].args to be a json array", index);
        return true;
    }
    size_t el_index;
    json_t* el;
    json_array_foreach(*args, el_index, el) {
        if (!json_is_string(el)) {
            log_err("expected element [%ld].args[%ld] to be a json string", index, el_index);
            return true;
        }
    }
    return false;
}

static bool validate_opts(const json_t* item, json_t** opts, const size_t index) {
    *opts = json_object_get(item, "opts");
    if (!*opts) {
        log_err("missing required prop [%ld].opts", index);
        return true;
    }
    if (*opts && !json_is_string(*opts)) {
        log_err("expected prop [%ld].opts to be a json string", index);
        return true;
    }
    return false;
}

static bool validate_lopts(const json_t* item, json_t** lopts, const size_t index) {
    *lopts = json_object_get(item, "lopts");
    if (!*lopts) {
        log_err("missing required prop [%ld].lopts", index);
        return true;
    }
    if (*lopts && !json_is_string(*lopts)) {
        log_err("expected prop [%ld].lopts to be a json string", index);
        return true;
    }
    return false;
}

typedef struct Input {
    size_t index;
    json_t* label;
    json_t* args;
    json_t* opts;
    json_t* lopts;
} Input;

static bool validate_case(json_t* item, const size_t index, Input* input) {
    if (!json_is_object(item)) {
        log_err("expected element [%ld] to be a json object", index);
        return true;
    }
    bool err = validate_label(item, &input->label, index);
    if (err) {
        return err;
    }
    err = validate_args(item, &input->args, index);
    if (err) {
        return err;
    }
    err = validate_opts(item, &input->opts, index);
    if (err) {
        return err;
    }
    err = validate_lopts(item, &input->lopts, index);
    if (err) {
        return err;
    }
    return false;
}

typedef enum GetoptMode {
    GETOPT_MODE_GNU,
    GETOPT_MODE_POSIX,
    GETOPT_MODE_INORDER,
    GETOPT_MODE_COUNT,
} GetoptMode;

const char* mode_prefixes[GETOPT_MODE_COUNT] = {
    [GETOPT_MODE_GNU] = ":",
    [GETOPT_MODE_POSIX] = "+:",
    [GETOPT_MODE_INORDER] = "-:",
};

const char* getopt_mode_name(GetoptMode mode) {
    switch (mode) {
        case GETOPT_MODE_POSIX:
            return "posix";
        case GETOPT_MODE_INORDER:
            return "inorder";
        default:
            return "gnu";
    }
}

typedef enum GetoptFunc {
    GETOPT_FUNC_GETOPT,
    GETOPT_FUNC_GETOPT_LONG,
    GETOPT_FUNC_GETOPT_LONG_ONLY,
    GETOPT_FUNC_COUNT,
} GetoptFunc;

const char* getopt_func_name(GetoptFunc func) {
    switch (func) {
        case GETOPT_FUNC_GETOPT_LONG:
            return "getopt_long";
        case GETOPT_FUNC_GETOPT_LONG_ONLY:
            return "getopt_long_only";
        default:
            return "getopt";
    }
}

static int lopts_count(const char* optstring) {
    char* dupstr = strdup(optstring);
    if (!dupstr) {
        log_err("\"%s\" while allocating longopts", strerror(errno));
        return -1;
    }

    int count = 0;
    char* token = strtok(dupstr, ",");
    while (token) {
        count++;
        token = strtok(NULL, ",");
    }
    free(dupstr);

    return count;
}

static struct option* alloc_lopts(const int count) {
    return calloc(count + 1, sizeof(struct option));  // +1 for NULL terminator
}

void lopts_destroy(struct option* lopts) {
    if (lopts) {
        for (int i = 0; lopts[i].name != NULL; i++) {
            free((char*)lopts[i].name);
        }
        free(lopts);
    }
}

static bool parse_lopts(const char* optstring, struct option* lopts) {
    char* dupstr = strdup(optstring);
    if (!dupstr) {
        log_err("\"%s\" while allocating longopts", strerror(errno));
        return true;  // allocation failed
    }

    int i = 0;
    char* token = strtok(dupstr, ",");
    while (token) {
        char* name = token;
        int has_arg = no_argument;

        size_t len = strlen(token);
        if (len > 0 && token[len - 1] == ':') {
            token[len - 1] = '\0';
            has_arg = required_argument;
            if (len > 1 && token[len - 2] == ':') {
                token[len - 2] = '\0';
                has_arg = optional_argument;
            }
        }

        lopts[i].name = strdup(name);
        lopts[i].has_arg = has_arg;
        lopts[i].flag = NULL;
        lopts[i].val = -2;

        if (!lopts[i].name) {
            log_err("\"%s\" while allocating longopts", strerror(errno));
            for (int j = 0; j < i; j++) {
                free((char*)lopts[j].name);
            }
            free(dupstr);
            return true;  // allocation failed
        }

        i++;
        token = strtok(NULL, ",");
    }

    free(dupstr);
    return false;
}

struct option* create_lopts(const char* optstring) {
    if (!optstring || *optstring == '\0') {
        return calloc(1, sizeof(struct option));
    }

    int count = lopts_count(optstring);
    if (count < 0) {
        return NULL;
    }

    struct option* lopts = alloc_lopts(count);
    if (!lopts) {
        return NULL;
    }

    if (parse_lopts(optstring, lopts)) {
        lopts_destroy(lopts);
        return NULL;
    }

    return lopts;
}

static char* trim_name(const char* input) {
    if (input == NULL) return NULL;

    while (*input == '-') {
        input++;
    }

    const char* equals = strchr(input, '=');

    size_t len = equals ? (size_t)(equals - input) : strlen(input);
    char* result = malloc(len + 1);
    if (result == NULL) {
        log_err("\"%s\" while allocating fixture name", strerror(errno));
        return NULL;
    }

    strncpy(result, input, len);
    result[len] = '\0';

    return result;
}

static bool handle_case(json_t* item, const size_t index, json_t* results_array) {
    Input input = {
        .label = NULL,
        .args = NULL,
        .opts = NULL,
        .lopts = NULL,
    };
    bool err = validate_case(item, index, &input);
    if (err) {
        return true;
    }

    for (GetoptFunc func = 0; func < GETOPT_FUNC_COUNT; func++) {
        for (GetoptMode mode = 0; mode < GETOPT_MODE_COUNT; mode++) {
            json_t* label = json_object_get(item, "label");
            json_t* args = json_object_get(item, "args");
            json_t* opts = json_object_get(item, "opts");
            json_t* lopts = json_object_get(item, "lopts");

            json_t* result = json_object();
            json_object_set(result, "label", label);
            json_object_set_new(result, "func", json_string(getopt_func_name(func)));
            json_object_set_new(result, "mode", json_string(getopt_mode_name(mode)));
            json_object_set(result, "args", args);

            json_t* iter_array = json_array();
            json_object_set_new(result, "want_results", iter_array);

            int argc = json_array_size(args);
            char** argv = calloc(argc + 1, sizeof(char*));
            if (!argv) {
                return true;
            }

            int i;
            json_t* arg;
            json_array_foreach(args, i, arg) {
                argv[i] = strdup(json_string_value(arg));
                if (!argv[i]) {
                    log_err("\"%s\" while allocating longopts", strerror(errno));
                    for (int j = 0; j < i; j++) {
                        free(argv[i]);
                    }
                    free(argv);
                    return true;
                }
            }

            size_t prefix_len = strlen(mode_prefixes[mode]);
            size_t opts_len = strlen(json_string_value(opts));
            char optstring[prefix_len + opts_len + 1];

            strcpy(optstring, mode_prefixes[mode]);
            strcat(optstring, json_string_value(opts));

            json_t* opts_array = json_array();
            i = 0;
            while (json_string_value(opts)[i] != '\0') {
                if (json_string_value(opts)[i] != ':') {
                    char element = json_string_value(opts)[i];
                    int colon_count = 0;

                    while (json_string_value(opts)[i + 1] == ':') {
                        colon_count++;
                        i++;
                    }

                    const char* has_arg;
                    switch (colon_count) {
                        case no_argument:
                            has_arg = "no_argument";
                            break;
                        case required_argument:
                            has_arg = "required_argument";
                            break;
                        case optional_argument:
                            has_arg = "optional_argument";
                            break;
                        default:
                            log_err("invalid optstring %s", json_string_value(opts));
                            for (i = 0; i < argc; i++) {
                                if (argv[i]) {
                                    free(argv[i]);
                                }
                            }
                            free(argv);
                            return true;
                    }

                    json_array_append_new(opts_array, json_pack("{s:i,s:s}", "char", element, "has_arg", has_arg));
                    i++;
                }
            }
            json_object_set_new(result, "opts", opts_array);

            struct option* longoptions = create_lopts(json_string_value(lopts));
            if (!longoptions) {
                for (i = 0; i < argc; i++) {
                    if (argv[i]) {
                        free(argv[i]);
                    }
                }
                free(argv);
                return true;
            }

            json_t* json_lopts_array = json_array();
            struct option* current_lopt = longoptions;
            while ((*current_lopt).name != NULL) {
                const char* has_arg = "no_argument";
                switch ((*current_lopt).has_arg) {
                    case required_argument:
                        has_arg = "required_argument";
                        break;
                    case optional_argument:
                        has_arg = "optional_argument";
                        break;
                    default:
                        break;
                }
                json_array_append_new(json_lopts_array, json_pack("{s:s,s:s}", "name", (*current_lopt).name, "has_arg", has_arg));
                current_lopt++;
            }
            json_object_set_new(result, "lopts", json_lopts_array);

            optind = 0;
            opterr = 0;
            optopt = 0;
            int opt;
            int longindex = 0;
            for (;;) {
                if (func == GETOPT_FUNC_GETOPT) {
                    opt = getopt(argc, argv, optstring);
                }
                if (func == GETOPT_FUNC_GETOPT_LONG) {
                    opt = getopt_long(argc, argv, optstring, longoptions, &longindex);
                }
                if (func == GETOPT_FUNC_GETOPT_LONG_ONLY) {
                    opt = getopt_long_only(argc, argv, optstring, longoptions, &longindex);
                }

                json_t* json_char = json_integer(0);
                json_t* json_name = json_string("");
                json_t* json_err = json_string("");
                switch (opt) {
                    case ':':
                        json_string_set(json_err, "missing_opt_arg");
                        if (optopt > 0) {
                            json_integer_set(json_char, optopt);
                        } else if (func != GETOPT_FUNC_GETOPT) {
                            char* name = trim_name(argv[optind - 1]);
                            if (name == NULL) {
                                for (i = 0; i < argc; i++) {
                                    if (argv[i]) {
                                        free(argv[i]);
                                    }
                                }
                                free(argv);
                                lopts_destroy(longoptions);
                                return true;
                            }
                            json_string_set(json_name, name);
                            free(name);
                        }
                        break;
                    case '?':
                        json_string_set(json_err, "unknown_opt");
                        if (optopt > 0) {
                            json_integer_set(json_char, optopt);
                        } else {
                            char* name = trim_name(argv[optind - 1]);
                            if (name == NULL) {
                                for (i = 0; i < argc; i++) {
                                    if (argv[i]) {
                                        free(argv[i]);
                                    }
                                }
                                free(argv);
                                lopts_destroy(longoptions);
                                return true;
                            }
                            json_string_set(json_name, name);

                            current_lopt = longoptions;
                            while ((*current_lopt).name != NULL) {
                                if (strcmp(name, current_lopt->name) == 0) {
                                    json_string_set(json_err, "illegal_opt_arg");
                                    char* inlineOptarg = strchr(argv[optind - 1], '=');
                                    if (inlineOptarg != NULL && *(inlineOptarg + 1) != '\0') {
                                        optarg = inlineOptarg + 1;
                                    }
                                    break;
                                }
                                current_lopt++;
                            }
                            free(name);
                        }
                        break;
                    case -1:
                        json_string_set(json_err, "done");
                        break;
                    case -2:
                        json_string_set(json_name, longoptions[longindex].name);
                        break;
                    default:
                        json_integer_set(json_char, opt);
                        break;
                }

                json_array_append_new(
                    iter_array,
                    json_pack(
                        "{s:o,s:o,s:s,s:o}",
                        "char", json_char,
                        "name", json_name,
                        "optarg", optarg ? optarg : "",
                        "err", json_err));

                longindex = 0;

                if (opt == -1) {
                    json_object_set_new(result, "want_optind", json_integer(optind));
                    json_t* want_args_array = json_array();
                    json_object_set_new(result, "want_args", want_args_array);
                    for (i = 0; i < argc; i++) {
                        json_array_append_new(want_args_array, json_string(argv[i]));
                    }

                    break;
                }
            }

            for (i = 0; i < argc; i++) {
                if (argv[i]) {
                    free(argv[i]);
                }
            }
            free(argv);
            lopts_destroy(longoptions);
            json_array_append_new(results_array, result);
        }
    }

    return false;
}

int main(void) {
    json_t* root;
    json_error_t json_err;

    root = json_load_file(INFILE_PATH, 0, &json_err);
    if (!root) {
        log_err("decoding %s, %s at line %d, col %d", INFILE_PATH, json_err.text, json_err.line, json_err.column);
        json_decref(root);
        return EXIT_FAILURE;
    }

    if (!json_is_array(root)) {
        log_err("expected input to be a json array");
        json_decref(root);
        return EXIT_FAILURE;
    }
    log_info("loaded %d cases", json_array_size(root));

    json_t* results_array = json_array();

    size_t index;
    json_t* item;
    json_array_foreach(root, index, item) {
        bool err = handle_case(item, index, results_array);
        if (err) {
            json_decref(results_array);
            json_decref(root);
            return EXIT_FAILURE;
        }
    }
    log_info("generated %d fixtures", json_array_size(results_array));

    int err = json_dump_file(results_array, OUTFILE_PATH, JSON_INDENT(4));
    if (err) {
        log_err("\"%s\" while opening %s", strerror(errno), OUTFILE_PATH);
        json_decref(results_array);
        json_decref(root);
        return EXIT_FAILURE;
    }

    json_decref(results_array);
    json_decref(root);
    return EXIT_SUCCESS;
}
