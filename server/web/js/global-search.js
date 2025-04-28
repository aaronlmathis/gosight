// public/js/global-search.js
document.addEventListener('DOMContentLoaded', () => {
    const input = document.getElementById('global-search');
    const results = document.getElementById('search-results');

    let controller = null; // for aborting previous fetch

    input.addEventListener('input', async (e) => {
        const term = e.target.value.trim();

        if (controller) controller.abort(); // cancel previous request
        controller = new AbortController();

        if (term.length < 2) {
            results.classList.add('hidden');
            results.innerHTML = '';
            return;
        }

        try {
            const res = await fetch(`/api/v1/search?term=${encodeURIComponent(term)}`, {
                signal: controller.signal,
            });

            if (!res.ok) {
                console.error('Search failed');
                return;
            }

            const data = await res.json();
            renderResults(data);
        } catch (err) {
            if (err.name !== 'AbortError') {
                console.error('Fetch error:', err);
            }
        }
    });

    function renderResults(data) {
        results.innerHTML = '';
        let found = false;

        const sections = [
            { key: 'endpoints', title: 'Endpoints' },
            { key: 'rules', title: 'Rules' },
            { key: 'tags', title: 'Tags' },
            { key: 'logs', title: 'Logs' },
        ];

        for (const section of sections) {
            if (data[section.key] && data[section.key].length > 0) {
                found = true;

                const header = document.createElement('li');
                header.className = 'px-3 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400';
                header.textContent = section.title;
                results.appendChild(header);

                for (const item of data[section.key]) {
                    const li = document.createElement('li');
                    li.className = 'px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer text-sm text-gray-800 dark:text-gray-200';
                    li.textContent = item.label;

                    // Attach click navigation
                    li.addEventListener('click', () => handleNavigation(section.key, item));
                    results.appendChild(li);
                }
            }
        }

        results.classList.toggle('hidden', !found);
    }

    function handleNavigation(section, item) {
        if (section === 'endpoints') {
            window.location.href = `/endpoints/${item.endpoint_id}`;
        } else if (section === 'rules') {
            window.location.href = `/rules/${item.rule_id}`;
        } else if (section === 'tags') {
            const [key, value] = item.label.split(':');
            window.location.href = `/endpoints?tag=${encodeURIComponent(key)}:${encodeURIComponent(value)}`;
        } else if (section === 'logs') {
            window.location.href = `/logs?search=${encodeURIComponent(item.label)}`;
        }
    }

    // Hide suggestions if clicking outside
    document.addEventListener('click', (e) => {
        if (!input.contains(e.target) && !results.contains(e.target)) {
            results.classList.add('hidden');
        }
    });

    // Hide suggestions on ESC
    input.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            results.classList.add('hidden');
        }
    });
});
