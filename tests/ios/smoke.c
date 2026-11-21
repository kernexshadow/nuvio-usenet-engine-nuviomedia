// Runs the iOS library inside a simulator process the way the app will: start,
// ask /health with and without the token, restart, stop.
// Usage: smoke CACHE_DIR
#include <arpa/inet.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <unistd.h>

#include "nuvio_usenet.h"

static const char *TOKEN = "0123456789abcdef0123456789abcdef";

static int start(const char *cache_dir) {
    char bootstrap[4096];
    snprintf(bootstrap, sizeof bootstrap,
             "{\"token\":\"%s\",\"certificates\":[],\"memoryMiB\":128,\"nzbCacheDir\":\"%s\"}", TOKEN, cache_dir);
    char *record = nuvio_usenet_start(bootstrap);
    printf("start: %s\n", record);
    const char *port = strstr(record, "\"port\":");
    int result = port ? atoi(port + 7) : 0;
    nuvio_usenet_free(record);
    return result;
}

// Returns the HTTP status of GET /health, or -1 when nothing answers.
static int health(int port, const char *token) {
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    struct sockaddr_in addr = {0};
    addr.sin_family = AF_INET;
    addr.sin_port = htons((unsigned short)port);
    addr.sin_addr.s_addr = htonl(INADDR_LOOPBACK);
    if (fd < 0 || connect(fd, (struct sockaddr *)&addr, sizeof addr) != 0) {
        if (fd >= 0) close(fd);
        return -1;
    }
    char request[256];
    int n = snprintf(request, sizeof request, "GET /health HTTP/1.0\r\n%s%s%s\r\n",
                     token ? "Authorization: Bearer " : "", token ? token : "", token ? "\r\n" : "");
    char response[512] = {0};
    int status = -1;
    if (write(fd, request, (size_t)n) == n && read(fd, response, sizeof response - 1) > 12) {
        status = atoi(response + 9);
    }
    close(fd);
    return status;
}

#define CHECK(cond, ...) do { if (!(cond)) { printf("FAIL: " __VA_ARGS__); printf("\n"); return 1; } } while (0)

int main(int argc, char **argv) {
    const char *cache_dir = argc > 1 ? argv[1] : "/tmp";
    int first = start(cache_dir);
    CHECK(first > 0, "no port");
    CHECK(health(first, TOKEN) == 200, "health with token");
    CHECK(health(first, NULL) == 401, "health without token");
    int second = start(cache_dir);
    CHECK(second > 0, "no port after restart");
    CHECK(first == second || health(first, TOKEN) == -1, "first server still answers");
    CHECK(health(second, TOKEN) == 200, "health after restart");
    nuvio_usenet_stop();
    CHECK(health(second, TOKEN) == -1, "server answers after stop");
    printf("ok\n");
    return 0;
}
