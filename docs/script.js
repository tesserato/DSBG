/**
 * Initializes tag filters by:
 * 1. Extracting unique tags from all button elements within the document.
 * 2. Sorting these tags alphabetically.
 * 3. Creating corresponding filter buttons for each tag.
 * 4. Implementing "Show All" and "Hide All" functionality for tag filtering.
 */
function initializeTagFilters() {
    // Create a Set to store unique tags. Using a set ensures each tag is only stored once.
    const tags = new Set();

    // Iterate through all button elements in the document.
    for (const tagElement of document.getElementsByTagName("button")) {
        // Add the trimmed inner HTML of each button to the tags Set.
        // This ensures that tags with extra whitespace are treated the same.
        tagElement.innerHTML = tagElement.innerHTML.trim();
        tags.add(tagElement.innerHTML);
    }

    // Convert the Set of tags to an Array and sort it alphabetically.
    const sortedTags = Array.from(tags);
    sortedTags.sort(Intl.Collator().compare);

    // Get the container element where the tag buttons will be placed.
    // This is the element with ID "buttons" in the HTML.
    const btnContainer = document.getElementById("buttons");

    // Create filter buttons for each tag.
    for (const tag of sortedTags) {
        const btn = document.createElement("button");
        btn.className = "on"; // Initially set the button to the 'on' state (selected).
        btn.type = "button";
        btn.innerHTML = tag;
        btnContainer.appendChild(btn); // Add the new button to the container.
    }

    // Create a "Show All" button.
    const showAllBtn = document.createElement("button");
    showAllBtn.className = "on"; // Initially set to on (selected).
    showAllBtn.innerHTML = "⬤"; // Use a circle character for visual representation.
    showAllBtn.id = "show_all_btn"; // Set the ID to be able to reference it easily.
    showAllBtn.title = "Select all tags"; // A title for accessibility.
    showAllBtn.type = "button";

    // Create a "Hide All" button.
    const hideAllBtn = document.createElement("button");
    hideAllBtn.className = "on"; // Initially set to on (selected).
    hideAllBtn.innerHTML = "⬤"; // Use a circle character for visual representation.
    hideAllBtn.id = "hide_all_btn"; // Set the ID to be able to reference it easily.
    hideAllBtn.title = "De-select all tags"; // A title for accessibility.
    hideAllBtn.type = "button";

    // Insert the "Show All" and "Hide All" buttons at the beginning of the container.
    btnContainer.insertBefore(hideAllBtn, btnContainer.firstChild);
    btnContainer.insertBefore(showAllBtn, btnContainer.firstChild);

    // Get all elements with the class 'detail' (the posts).
    // These are the main article containers, the filter will be applied to them.
    const posts = document.getElementsByClassName('detail');

    // Create a Set to store the tag filter buttons (excluding "Show All" and "Hide All").
    // This will be used for filtering the displayed posts.
    const filterButtons = new Set();
    for (const btn of document.getElementsByTagName('button')) {
        if (btn.id !== "show_all_btn" && btn.id !== "hide_all_btn") {
            filterButtons.add(btn);
        }
    }

    /**
     * Refreshes the visibility of posts based on the currently active tag filters.
     * It iterates through each post, checks its tag buttons, and sets its display
     * to "block" (visible) if any tag is "on" and "none" (hidden) if all are "off".
     */
    function refreshPosts() {
        for (const post of posts) {
            const buttons = post.getElementsByTagName("button");
            let isVisible = false;

            for (const btn of buttons) {
                if (btn.classList.contains("on")) {
                    isVisible = true;
                    break;
                }
            }
            post.style.display = isVisible ? "" : "none";
        }
    }

    // Add event listeners to the tag filter buttons.
    for (const btn of filterButtons) {
        btn.addEventListener("click", function (e) {
            // Check if all filter buttons were 'on' before the click.
            let allButtonsOn = true;
            for (const filterBtn of filterButtons) {
                if (filterBtn.className === "off") {
                    allButtonsOn = false;
                    break;
                }
            }

            const target = e.target; // Get the clicked button.

            // If all buttons were on, turn all off except the clicked one.
            // This allows to select only one tag at a time, simplifying the filtering.
            if (allButtonsOn) {
                for (const filterBtn of filterButtons) {
                    filterBtn.className = "off";
                }
                target.className = "on";
            } else {
                // Otherwise, toggle the state of the clicked button (on to off, or off to on).
                target.className = target.className === "on" ? "off" : "on";
            }

            // Ensure consistency between the clicked button and other buttons with the same tag.
            const targetInner = target.innerHTML;
            for (const filterBtn of filterButtons) {
                if (targetInner === filterBtn.innerHTML) {
                    filterBtn.className = target.className;
                }
            }
            refreshPosts(); // Update the visibility of posts.
        }, false);
    }

    // Add event listener to the "Show All" button.
    showAllBtn.addEventListener("click", function () {
        // Set all filter buttons to 'on' to show all posts.
        for (const btn of filterButtons) {
            btn.className = "on";
        }
        refreshPosts(); // Update the visibility of posts.
    }, false);

    // Add event listener to the "Hide All" button.
    hideAllBtn.addEventListener("click", function () {
        // Set all filter buttons to 'off' to hide all posts.
        for (const btn of filterButtons) {
            btn.className = "off";
        }
        refreshPosts(); // Update the visibility of posts.
    }, false);
}

// Call the function to initialize the tag filters when the script runs.
initializeTagFilters();

/**
 * Copies a generated Markdown summary of an article to the clipboard.
 * Reads article data from the data-* attributes of the clicked element.
 * Does not provide visual feedback on success to keep UI clean.
 * @param {HTMLElement} element The button element that was clicked.
 */
function copyMarkdownToClipboard(element) {
    // 1. Read all the data from the element's data attributes.
    const title = element.dataset.title;
    const description = element.dataset.description;
    const articleUrl = element.dataset.url;
    const imageUrl = element.dataset.imageUrl;
    const tags = element.dataset.tags ? element.dataset.tags.split(',') : [];

    // Check if the full raw markdown content is available via a sibling hidden textarea
    let rawText = null;
    try {
        // Look for the textarea specifically to use its .value property
        const rawTextEl = element.parentNode.querySelector('textarea.dsbg-raw-text');
        if (rawTextEl) {
            rawText = rawTextEl.value; // .value handles HTML entity decoding automatically
        }
    } catch (e) {
        console.warn("Could not retrieve raw markdown text", e);
    }

    let markdownString = "";

    if (rawText) {
        // Use the raw markdown content if available
        markdownString = rawText;
    } else {
        // Fallback: Build the Markdown string in the summary format.
        markdownString = `# ${title}\n\n`;

        if (description) {
            markdownString += `${description}\n\n`;
        }

        if (imageUrl) {
            markdownString += `![${title}](${imageUrl})\n\n`;
        }

        markdownString += `${articleUrl}\n\n`;

        if (tags.length > 0) {
            const hashtags = tags.map(tag => `#${tag.trim().replace(/ /g, '')}`).join(' ');
            markdownString += `${hashtags}`;
        }
    }

    // 3. Use the modern Navigator API to copy to the clipboard.
    // We do not modify the DOM/UI on success.
    navigator.clipboard.writeText(markdownString).catch(err => {
        // 5. Log an error if copying fails.
        console.error('Failed to copy Markdown: ', err);
        alert('Failed to copy Markdown. See console for details.');
    });
}