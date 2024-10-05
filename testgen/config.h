#ifndef CONFIG_H
#define CONFIG_H

typedef struct Config {
    const char *inpath;
    const char *outpath;
} Config;

Config *create_config(int argc, char *argv[]);
void config_destroy(Config *cfg);

#endif
