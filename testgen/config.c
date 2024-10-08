
#include "config.h"

#include <errno.h>
#include <getopt.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void print_usage(const char *name);

Config *create_config(int argc, char *argv[]) {
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

void config_destroy(Config *cfg) {
    if (cfg != NULL) {
        free((char *)cfg->outpath);
        free((char *)cfg->inpath);
        free(cfg);
    }
}

static void print_usage(const char *name) {
    fprintf(stderr, "usage: %s -o <outfile> <infile>\n", name);
}
