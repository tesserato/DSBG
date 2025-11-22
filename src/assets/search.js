document.addEventListener('DOMContentLoaded', function () {
    const searchInput = document.getElementById('search-input');
    const searchResults = document.getElementById('search-results');
    const articlesContainer = document.getElementById('articles-container');
    let lunrIndex;
    let articleMap = {};
    const minSearchChars = 3; // Minimum characters to trigger a search

    // Initialize search by fetching the index
    async function initializeSearch() {
        try {
            const response = await fetch('search_index.json');
            if (!response.ok) throw new Error('Network response was not ok.');
            const articleData = await response.json();

            // Map articles by URL (which serves as the ID/Ref)
            articleData.forEach(article => {
                articleMap[article.url] = article;
            });

            // Build the Lunr index
            lunrIndex = lunr(function () {
                this.ref('url');
                this.field('title', { boost: 10 });
                this.field('description');
                this.field('tags', { boost: 5 });
                this.field('content'); // Index Plain Text content

                articleData.forEach(doc => this.add(doc));
            });
        } catch (error) {
            console.error("Failed to initialize search:", error);
            searchInput.placeholder = "Search index failed to load.";
            searchInput.disabled = true;
        }
    }

    // Helper: Build a fuzzy query for Lunr
    function buildFuzzyQuery(term) {
        const terms = term.trim().split(/\s+/).filter(word => word.length > 0);
        if (terms.length === 0) return '';

        return terms.map((word, i) => {
            let fuzzyWord = word + '~1'; // Fuzzy match with distance 1
            // Add a wildcard to the last term for "type-ahead" feel
            if (i === terms.length - 1) {
                fuzzyWord = word + '*' + ' ' + fuzzyWord;
            }
            return fuzzyWord;
        }).join(' ');
    }

    // Helper: Create a highlighted snippet from Plain Text content
    function createSnippet(content, term) {
        // Content is now plain text, so no HTML stripping needed.
        // Just normalize whitespace.
        const textContent = content.replace(/\s+/g, ' ');
        const termIndex = textContent.toLowerCase().indexOf(term.toLowerCase());

        // If term not found (e.g. fuzzy match), just return start of text
        if (termIndex === -1) return (textContent.substring(0, 150) + '...');

        const start = Math.max(0, termIndex - 50);
        const end = Math.min(textContent.length, termIndex + term.length + 100);
        let snippet = textContent.substring(start, end);

        if (start > 0) snippet = '...' + snippet;
        if (end < textContent.length) snippet = snippet + '...';

        // Escape HTML special chars in snippet to prevent injection
        snippet = snippet.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

        // Simple highlighting logic
        const highlightTerm = term.split(/\s+/).pop().replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&');
        return snippet.replace(new RegExp(`(${highlightTerm})`, 'gi'), '<strong>$1</strong>');
    }

    // Event Listener: Keyup on search input
    searchInput.addEventListener('keyup', () => {
        if (!lunrIndex) return;
        const term = searchInput.value.trim();

        // If query is too short, hide results and show original articles
        if (!term || term.length < minSearchChars) {
            articlesContainer.style.display = 'block';
            searchResults.style.display = 'none';
            searchResults.innerHTML = '';
            return;
        }

        try {
            const fuzzyQuery = buildFuzzyQuery(term);
            const results = lunrIndex.search(fuzzyQuery);

            // Hide article list, show search results
            articlesContainer.style.display = 'none';
            searchResults.style.display = 'block';

            if (results.length === 0) {
                searchResults.innerHTML = '<li>No results found.</li>';
            } else {
                searchResults.innerHTML = results.map(result => {
                    const article = articleMap[result.ref];
                    // Use content (plain text) for snippet generation
                    const snippet = createSnippet(article.content, term);
                    // Use article.url (the ref) for the link
                    return `<li><a href="${article.url}">${article.title}</a><div class="search-result-snippet">${snippet}</div></li>`;
                }).join('');
            }
        } catch (e) {
            // Lunr might throw on invalid query syntax
            searchResults.innerHTML = '<li>Invalid search query.</li>';
        }
    });

    // Event Listener: Clear results when clicking a link or if necessary
    searchResults.addEventListener('click', (e) => {
        if (e.target.tagName === 'A') {
            // Optional: Clear search when navigating away (useful for SPA, less for multi-page but good cleanup)
        }
    });

    initializeSearch();
});