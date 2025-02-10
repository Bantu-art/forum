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
// Add these functions to your like.js file

function toggleEditComment(commentId) {
    const contentDiv = document.getElementById(`comment-content-${commentId}`);
    const editForm = document.getElementById(`comment-edit-${commentId}`);
    
    if (contentDiv.style.display !== 'none') {
        contentDiv.style.display = 'none';
        editForm.style.display = 'block';
    } else {
        contentDiv.style.display = 'block';
        editForm.style.display = 'none';
    }
}

function handleEditComment(event, commentId, postId) {
    event.preventDefault();
    
    const form = event.target;
    const content = form.querySelector('textarea').value;

    fetch('/comment/edit', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: `comment_id=${commentId}&post_id=${postId}&content=${encodeURIComponent(content)}`,
        credentials: 'include'
    })
    .then(response => {
        if (response.ok) {
            window.location.reload();
        } else {
            throw new Error('Failed to edit comment');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert(error.message);
    });
}

function deleteComment(commentId, postId) {
    if (!confirm('Are you sure you want to delete this comment?')) {
        return;
    }

    fetch('/comment/delete', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: `comment_id=${commentId}&post_id=${postId}`,
        credentials: 'include'
    })
    .then(response => {
        if (response.ok) {
            window.location.reload();
        } else {
            throw new Error('Failed to delete comment');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert(error.message);
    });
}