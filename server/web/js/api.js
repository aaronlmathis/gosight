export function gosightFetch(url, options = {}) {
    return fetch(url, {
        credentials: "same-origin",
        headers: {
            "Content-Type": "application/json",
            ...options.headers,
        },
        method: "GET",
        ...options,
    });
}