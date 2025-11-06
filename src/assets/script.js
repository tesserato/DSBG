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
        btn.innerHTML = tag;
        btnContainer.appendChild(btn); // Add the new button to the container.
    }

    // Create a "Show All" button.
    const showAllBtn = document.createElement("button");
    showAllBtn.className = "on"; // Initially set to on (selected).
    showAllBtn.innerHTML = "⬤"; // Use a circle character for visual representation.
    showAllBtn.id = "show_all_btn"; // Set the ID to be able to reference it easily.
    showAllBtn.title = "Select all tags"; // A title for accessibility.

    // Create a "Hide All" button.
    const hideAllBtn = document.createElement("button");
    hideAllBtn.className = "on"; // Initially set to on (selected).
    hideAllBtn.innerHTML = "⬤"; // Use a circle character for visual representation.
    hideAllBtn.id = "hide_all_btn"; // Set the ID to be able to reference it easily.
    hideAllBtn.title = "De-select all tags"; // A title for accessibility.

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
            let isVisible = false;
            // Iterate through the tag buttons within each post.
            for (const btn of post.getElementsByTagName("button")) {
                // If any tag button within the post is in the 'on' state, show the post.
                if (btn.className === "on") {
                    post.style.display = "";
                    break; // No need to check further tags for this post.
                }
                // If all tag buttons within the post are in the 'off' state, hide the post.
                post.style.display = "none";
            }
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
    showAllBtn.addEventListener("click", function (e) {
        // Set all filter buttons to 'on' to show all posts.
        for (const btn of filterButtons) {
            btn.className = "on";
        }
        refreshPosts(); // Update the visibility of posts.
    }, false);

    // Add event listener to the "Hide All" button.
    hideAllBtn.addEventListener("click", function (e) {
         // Set all filter buttons to 'off' to hide all posts.
        for (const btn of filterButtons) {
            btn.className = "off";
        }
        refreshPosts(); // Update the visibility of posts.
    }, false);
}

// Call the function to initialize the tag filters when the script runs.
initializeTagFilters();