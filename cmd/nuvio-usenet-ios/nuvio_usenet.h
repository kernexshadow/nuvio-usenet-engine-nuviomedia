#ifndef NUVIO_USENET_H
#define NUVIO_USENET_H

#ifdef __cplusplus
extern "C" {
#endif

/*
 * Starts the engine's loopback HTTP server inside the calling process, stopping
 * the running one first. bootstrap_json is the record the Android sidecar reads
 * from stdin: {"token": "...", "certificates": [...], "memoryMiB": N,
 * "nzbCacheDir": "..."}. Returns {"protocol":1,"port":N} or {"error":"..."};
 * release it with nuvio_usenet_free.
 */
char *nuvio_usenet_start(char *bootstrap_json);

/* Closes the server and its sessions; returns once they are closed. */
void nuvio_usenet_stop(void);

/* Releases a record returned by nuvio_usenet_start. */
void nuvio_usenet_free(char *record);

#ifdef __cplusplus
}
#endif

#endif
