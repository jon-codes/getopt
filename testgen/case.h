#ifndef CASE_H
#define CASE_H

#include <stdbool.h>
#include <stdio.h>

typedef struct Case {
    char *label;
    char *opts;
    char *lopts;
    char **argv;
    int argc;
} Case;

bool read_case(Case *dest, FILE *src);
void case_clear(Case *c);

#endif
