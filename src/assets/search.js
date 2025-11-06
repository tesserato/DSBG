document.addEventListener('DOMContentLoaded', function () {
    // Get references to the search input and results elements.
    const searchInput = document.getElementById('search-input');
    const searchResults = document.getElementById('search-results');
    // Declare a variable to store the Fuse.js instance.
    let fuse;

    // Fetch the search index from the JSON file.
    fetch('search_index.json')
        .then(response => response.json())
        .then(searchIndex => {
            // Configure Fuse.js options for searching.
            const options = {
                includeScore: true,  // Include the score of each match.
                findAllMatches: true, // Find all matches for a given search term.
                includeMatches: true, // Include information about matches (locations, etc)
                ignoreLocation: true, // Ignore the position of the match in the text.
                minMatchCharLength: 2, // Minimum number of characters needed to match.
                useExtendedSearch: true, // Enable the extended search capabilities.
                threshold: 0.35, // Set the threshold for the search. Lower = more relaxed matches.
                distance: 10,   // How close a match has to be, in characters.
                keys: ['title', 'content', 'description', 'tags'] // Keys in index to search for.
            };
            // Create a new Fuse.js instance with the search index and options.
            fuse = new Fuse(searchIndex, options);
        });

    // Add an event listener to the search input for 'keyup' events.
    // This triggers the search each time a key is released.
    searchInput.addEventListener('keyup', function (event) {
        // If Fuse.js is initialized (data is loaded)
        if (fuse) {
            const searchTerm = event.target.value; // Get the user's search term.
            // Perform the search using Fuse.js.
            const results = fuse.search(searchTerm);

            // Prepare the HTML to display the search results.
            let resultsHTML = '';
            // If there are no results, show the "No results found" message.
            if (results.length === 0) {
                resultsHTML = '<li>No results found.</li>';
            } else {
               // If results are found, iterate through them to generate an HTML list.
               // The score is used to sort results, and a link is created for each result.
                results.forEach(result => {
                    const article = result.item;
                    resultsHTML += `<li>${(1.0 - result.score).toFixed(2)} <a href="${article.url}">${article.title}</a></li>`;
                });
            }
            // Update the search results element with the generated HTML.
            searchResults.innerHTML = resultsHTML;
        }

    });

     // Prevent default behavior for mousedown events on search results.
    // This avoids focus stealing from the search input and unwanted selections.
    searchResults.addEventListener('mousedown', function (event) {
        event.preventDefault();
    });
    
    // Clear search results when input loses focus.
    searchInput.addEventListener('blur', function (event) {
        searchInput.value = ''; // Clear the search input.
        searchResults.innerHTML = ''; // Clear the result list.
    });
});