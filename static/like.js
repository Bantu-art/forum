function handleReaction(event) {
    const button = event.currentTarget;
    const postID = button.getAttribute("data-post-id"); // Get post_id from the button
    const action = button.getAttribute("data-action"); // "like" or "dislike"

    // Determine the reaction type (1 for like, 0 for dislike)
    const like = action === "like" ? 1 : 0;

    // Send a POST request to the /react endpoint
    fetch("/react", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            post_id: parseInt(postID), // Ensure postID is an integer
            like: like,
        }),
    })
    .then(response => {
        if (!response.ok) {
            // If the response is not OK, parse the error message
            return response.json().then(errorData => {
                throw new Error(errorData.error || "An error occurred");
            });
        }
        return response.json();
    })
    .then(data => {
        // Update the like and dislike counts in the UI
        const likesElement = document.getElementById(`likes-${postID}`);
        const dislikesElement = document.getElementById(`dislikes-${postID}`);

        if (likesElement && dislikesElement) {
            likesElement.textContent = data.likes;
            dislikesElement.textContent = data.dislikes;
        }
    })
    .catch(error => {
        console.error("Error:", error);
        alert(error.message); // Show the error message to the user
    });
}

// Add event listeners to all like/dislike buttons
document.querySelectorAll(".like-btn, .dislike-btn").forEach(button => {
    button.addEventListener("click", handleReaction);
});