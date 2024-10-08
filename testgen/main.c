#define _POSIX_C_SOURCE 200809L

#include <errno.h>
#include <getopt.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef enum Result {
    RESULT_OK,
    RESULT_ERR,
    RESULT_DONE,
} Result;

static void print_usage(const char *name) {
    fprintf(stderr, "usage: %s -o <outfile> <infile>\n", name);
}

typedef struct Config {
    const char *inpath;
    const char *outpath;
} Config;

static void config_destroy(Config *cfg) {
    if (cfg != NULL) {
        free((char *)cfg->outpath);
        free((char *)cfg->inpath);
        free(cfg);
    }
}

static Config *create_config(int argc, char *argv[]) {
    Config *cfg = calloc(1, sizeof(Config));
    if (cfg == NULL) {
        fprintf(stderr, "error allocating config: %s\n", strerror(errno));
        return NULL;
    }

    int opt;
    while ((opt = getopt(argc, argv, ":o:")) != -1) {
        switch (opt) {
            case 'o':
                cfg->outpath = strdup(optarg);
                if (cfg->outpath == NULL) {
                    fprintf(stderr, "error allocating config: %s\n", strerror(errno));
                    config_destroy(cfg);
                    return NULL;
                }
                break;
            case '?':
                fprintf(stderr, "error: Unknown option \"%c\"\n", optopt);
                print_usage(argv[0]);
                config_destroy(cfg);
                return NULL;
            case ':':
                fprintf(stderr, "error: Option \"%c\" requires an argument\n", optopt);
                print_usage(argv[0]);
                config_destroy(cfg);
                return NULL;
            default:
                break;
        }
    }

    if (cfg->outpath == NULL) {
        fprintf(stderr, "error: Option -o is required\n");
        print_usage(argv[0]);
        config_destroy(cfg);
        return NULL;
    }

    if (optind < argc) {
        cfg->inpath = strdup(argv[optind]);
        if (cfg->inpath == NULL) {
            fprintf(stderr, "error allocating config: %s\n", strerror(errno));
            config_destroy(cfg);
            return NULL;
        }
    } else {
        fprintf(stderr, "error: missing required infile parameter\n");
        print_usage(argv[optind]);
        config_destroy(cfg);
        return NULL;
    }

    return cfg;
}

typedef struct Case {
    char *label;
    char *opts;
    char *lopts;
    char **argv;
    int argc;
} Case;

static void case_clear(Case *c) {
    if (c != NULL) {
        free(c->label);
        free(c->opts);
        free(c->lopts);
        if (c->argv != NULL) {
            for (int i = 0; i < c->argc; i++) {
                free(c->argv[i]);
            }
            free(c->argv);
        }
    }
    c->label = NULL;
    c->opts = NULL;
    c->lopts = NULL;
    c->argv = NULL;
    c->argc = 0;
}

typedef struct CaseIterator {
    FILE *f;
    Case current;
    Result last_result;
} CaseIterator;

static void case_iterator_destroy(CaseIterator *iter) {
    if (iter != NULL) {
        if (iter->f != NULL) {
            fclose(iter->f);
        }
        case_clear(&iter->current);
        free(iter);
    }
}

static CaseIterator *create_case_iterator(Config *cfg) {
    CaseIterator *iter = calloc(1, sizeof(CaseIterator));
    if (iter == NULL) {
        fprintf(stderr, "error allocating case iterator: %s\n", strerror(errno));
        return NULL;
    }

    iter->f = fopen(cfg->inpath, "r");
    if (iter->f == NULL) {
        fprintf(stderr, "error opening infile %s: %s\n", cfg->inpath, strerror(errno));
        case_iterator_destroy(iter);
        return NULL;
    }

    return iter;
}

static bool iterator_has_next(CaseIterator *iter) {
    return iter != NULL && iter->last_result == RESULT_OK;
}

static Case *iterator_next(CaseIterator *iter) {
    if (!iterator_has_next(iter)) {
        // *item = NULL;
        // return iter->last_result;
    }
    return NULL
}

int main(int argc, char *argv[]) {
    Config *cfg = create_config(argc, argv);
    if (cfg == NULL) {
        return EXIT_FAILURE;
    }

    CaseIterator *iter = create_case_iterator(cfg);
    if (iter == NULL) {
        config_destroy(cfg);
        return EXIT_FAILURE;
    }

    for (;;) {
        Case *res = iterator_next(iter);
    }

    case_iterator_destroy(iter);
    config_destroy(cfg);
    return EXIT_SUCCESS;
}
