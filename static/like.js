function handleReaction(event) {
    event.stopPropagation(); // Prevent post link click when clicking like/dislike
    
    const button = event.currentTarget;
    const postID = button.getAttribute("data-post-id");
    const action = button.getAttribute("data-action");
    
    // Check if user is logged in by looking for session cookie
    const hasSession = document.cookie.includes('session_token=');
    if (!hasSession) {
        window.location.href = '/signin';
        return;
    }

    const like = action === "like" ? 1 : 0;

    fetch("/react", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            post_id: parseInt(postID),
            like: like,
        }),
        credentials: 'include' // Important: Send cookies with request
    })
    .then(response => {
        if (response.status === 401) {
            // Redirect to login if unauthorized
            window.location.href = '/signin';
            throw new Error('Please log in to react to posts');
        }
        return response.json();
    })
    .then(data => {
        if (data.error) {
            throw new Error(data.error);
        }
        const likesElement = document.getElementById(`likes-${postID}`);
        const dislikesElement = document.getElementById(`dislikes-${postID}`);
        
        if (likesElement && dislikesElement) {
            likesElement.textContent = data.likes;
            dislikesElement.textContent = data.dislikes;
        }
    })
    .catch(error => {
        console.error("Error:", error);
        alert(error.message);
    });
}

// Add event listeners to all like/dislike buttons
document.querySelectorAll(".like-btn, .dislike-btn").forEach(button => {
    button.addEventListener("click", handleReaction);
});