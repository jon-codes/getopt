#ifndef ITERATE_H
#define ITERATE_H

#include <stdio.h>

#include "case.h"
#include "config.h"

typedef enum IteratorStatus {
    ITER_OK,
    ITER_ERROR,
    ITER_DONE,
} IteratorStatus;

typedef struct Iterator {
    FILE *src;
    Case current;
    IteratorStatus status;
    int index;
} Iterator;

Iterator *create_iterator(Config *cfg);
void iterator_destroy(Iterator *iter);
void iterator_next(Iterator *iter);

#endif
