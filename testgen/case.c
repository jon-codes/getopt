#include "case.h"

#include <jansson.h>
#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

void case_clear(Case *c) {
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

        c->label = NULL;
        c->opts = NULL;
        c->lopts = NULL;
        c->argv = NULL;
        c->argc = 0;
    }
}

bool read_case(Case *dest, FILE *src) {
    size_t flags = JSON_DISABLE_EOF_CHECK;
    json_t *json_value;
    json_error_t json_err;

    json_value = json_loadf(src, flags, &json_err);
    if (!json_value) {
        fprintf(stderr, "JSON parsing error: %s\n", json_err.text);
        return true;
    }

    json_t *args;
    const char *label, *opts, *lopts;
    int err = json_unpack_ex(json_value, &json_err, 0, "{s:s, s:o, s:s, s:s}",
                             "label", &label,
                             "args", &args,
                             "opts", &opts,
                             "lopts", &lopts);

    if (err != 0) {
        fprintf(stderr, "JSON parsing error: %s %s\n", json_err.text, json_err.source);
        json_decref(json_value);
        return true;
    }

    if (!json_is_array(args)) {
        fprintf(stderr, "expected args to be an array\n");
        json_decref(json_value);
        return true;
    }

    dest->label = strdup(label);
    dest->opts = strdup(opts);
    dest->lopts = strdup(lopts);

    if (!dest->label || !dest->opts || !dest->lopts) {
        fprintf(stderr, "Memory allocation error\n");
        free(dest->label);
        free(dest->opts);
        free(dest->lopts);
        json_decref(json_value);
        return true;
    }

    dest->argc = json_array_size(args);
    dest->argv = malloc(dest->argc * sizeof(char *));
    if (dest->argv == NULL) {
        fprintf(stderr, "Memory allocation error for args\n");
        free(dest->label);
        free(dest->opts);
        free(dest->lopts);
        json_decref(json_value);
        return true;
    }

    for (int i = 0; i < dest->argc; i++) {
        json_t *el = json_array_get(args, i);
        if (!json_is_string(el)) {
            fprintf(stderr, "expected args element to be a string\n");
            for (int j = 0; j < i; j++) {
                free(dest->argv[j]);
            }
            free(dest->argv);
            free(dest->label);
            free(dest->opts);
            free(dest->lopts);
            json_decref(json_value);
            return true;
        }
        dest->argv[i] = strdup(json_string_value(el));
        if (dest->argv[i] == NULL) {
            fprintf(stderr, "Memory allocation error for arg %d\n", i);
            for (int j = 0; j < i; j++) {
                free(dest->argv[j]);
            }
            free(dest->argv);
            free(dest->label);
            free(dest->opts);
            free(dest->lopts);
            json_decref(json_value);
            return true;
        }
    }

    json_decref(json_value);
    return false;
}
