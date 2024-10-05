#include "iterate.h"

#include <ctype.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "case.h"
#include "config.h"
#include "stdbool.h"

static bool seek_array_start(FILE *f);
static enum IteratorStatus seek_next_element(FILE *f);
static FILE *prepare_infile(const Config *cfg);

Iterator *create_iterator(Config *cfg) {
    Iterator *iter = calloc(1, sizeof(Iterator));
    if (iter == NULL) {
        fprintf(stderr, "error allocating iterator: %s\n", strerror(errno));
        return NULL;
    }

    iter->src = prepare_infile(cfg);
    if (iter->src == NULL) {
        free(iter);
        return NULL;
    }

    iter->status = ITER_OK;
    iter->index = -1;

    return iter;
}

void iterator_destroy(Iterator *iter) {
    if (iter != NULL) {
        if (iter->src != NULL) {
            fclose(iter->src);
        }
        case_clear(&iter->current);
        free(iter);
    }
}

void iterator_next(Iterator *iter) {
    if (iter->status != ITER_OK) {
        return;
    }

    if (iter->index != -1) {
        iter->status = seek_next_element(iter->src);
        if (iter->status != ITER_OK) {
            return;
        }
    }

    case_clear(&iter->current);
    if (read_case(&iter->current, iter->src)) {
        iter->status = ITER_ERROR;
        return;
    }

    iter->index++;
}

static bool seek_array_start(FILE *f) {
    for (;;) {
        int c = fgetc(f);
        if (c == EOF || !isspace(c)) {
            if (c != '[') {
                fprintf(stderr, "error: Input is not a valid json array\n");
                return true;
            }
            break;
        }
    }
    return false;
}

static enum IteratorStatus seek_next_element(FILE *f) {
    if (f == NULL) {
        fprintf(stderr, "error: Invalid file pointer\n");
        return ITER_ERROR;
    }

    int c;
    while ((c = fgetc(f)) != EOF) {
        switch (c) {
            case ']':
                return ITER_DONE;
            case ',':
                return ITER_OK;
            default:
                if (!isspace(c)) {
                    fprintf(stderr, "error: Input is not a valid json array\n");
                    return ITER_ERROR;
                }
        }
    }
    fprintf(stderr, "error: Unexpected EOF\n");
    return ITER_ERROR;
}

static FILE *prepare_infile(const Config *cfg) {
    FILE *f = fopen(cfg->inpath, "r");
    if (f == NULL) {
        fprintf(stderr, "error opening iterator source %s: %s\n", cfg->inpath, strerror(errno));
        return NULL;
    }

    if (seek_array_start(f) != 0) {
        return NULL;
    }

    return f;
}
