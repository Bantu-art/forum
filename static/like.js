function handleReaction(event) {
    event.preventDefault();

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
        credentials: 'include'
    })
    .then(response => {
        if (response.status === 401) {

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

// Handler for comment reactions (modified to mirror the post reaction handler)
function handleCommentReaction(event) {
    event.stopPropagation(); // Prevent any unwanted propagation

    const button = event.currentTarget;
    const commentID = button.getAttribute("data-comment-id");
    const action = button.getAttribute("data-action");

    // Check if user is logged in by looking for session cookie
    const hasSession = document.cookie.includes('session_token=');
    if (!hasSession) {
        window.location.href = '/signin';
        return;
    }

    const like = action === "like" ? 1 : 0;

    fetch("/commentreact", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            comment_id: parseInt(commentID),
            like: like,
        }),
        credentials: 'include' // Ensure cookies are sent
    })
    .then(response => {
        if (response.status === 401) {
            window.location.href = '/signin';
            throw new Error('Please log in to react to comments');
        }
        return response.json();
    })
    .then(data => {
        if (data.error) {
            throw new Error(data.error);
        }
        // Update the comment's like and dislike counts on the page
        const likesElement = document.getElementById(`comment-likes-${commentID}`);
        const dislikesElement = document.getElementById(`comment-dislikes-${commentID}`);
        console.log(likesElement)
        
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

// Attach event listeners for post reaction buttons
document.querySelectorAll(".like-btn, .dislike-btn").forEach(button => {
    button.addEventListener("click", handleReaction);
});

// Attach event listeners for comment reaction buttons
document.querySelectorAll(".comment-like-btn, .comment-dislike-btn").forEach(button => {
    button.addEventListener("click", handleCommentReaction);
});
