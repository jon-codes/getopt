#define _POSIX_C_SOURCE 200809L

#include <stdlib.h>

#include "config.h"
#include "iterate.h"

int main(int argc, char *argv[]) {
    Config *cfg = create_config(argc, argv);
    if (cfg == NULL) {
        return EXIT_FAILURE;
    }

    Iterator *iter = create_iterator(cfg);
    if (iter == NULL) {
        config_destroy(cfg);
        return EXIT_FAILURE;
    }

    iterator_next(iter);
    for (; iter->status == ITER_OK; iterator_next(iter)) {
        printf("[%d].label: \"%s\"\n", iter->index, iter->current.label);
    }

    if (iter->status == ITER_ERROR) {
        iterator_destroy(iter);
        config_destroy(cfg);
        return EXIT_FAILURE;
    }

    iterator_destroy(iter);
    config_destroy(cfg);
    return EXIT_SUCCESS;
}
