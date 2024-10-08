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
} Case;

void case_destroy(Case *c) {
    if (c != NULL) {
        free(c->label);
    }
}

typedef struct CaseIterator {
    bool has_next;
} CaseIterator;

CaseIterator *create_case_iterator(Config *cfg) {
    return NULL;
}

int main(int argc, char *argv[]) {
    Config *cfg = create_config(argc, argv);
    if (cfg == NULL) {
        return EXIT_FAILURE;
    }

    FILE *infile = fopen(cfg->inpath, "r");
    if (infile == NULL) {
        fprintf(stderr, "error opening %s: %s\n", cfg->inpath, strerror(errno));
        config_destroy(cfg);
        return EXIT_FAILURE;
    }

    FILE *outfile = fopen(cfg->outpath, "w");
    if (outfile == NULL) {
        fprintf(stderr, "error opening %s: %s\n", cfg->inpath, strerror(errno));
        fclose(infile);
        config_destroy(cfg);
        return EXIT_FAILURE;
    }

    // CaseIterator *iter = create_case_iterator(cfg);
    // if (iter == NULL) {
    //     config_destroy(cfg);
    //     return EXIT_FAILURE;
    // }

    fclose(infile);
    fclose(outfile);
    config_destroy(cfg);
    return EXIT_SUCCESS;
}
