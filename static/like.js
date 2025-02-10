function handleReaction(event) {
    event.preventDefault();

    event.stopPropagation();

    const button = event.currentTarget;
    const postID = button.getAttribute("data-post-id");
    const action = button.getAttribute("data-action");

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

// Handler for comment reactions 
function handleCommentReaction(event) {
    event.stopPropagation(); 

    const button = event.currentTarget;
    const commentID = button.getAttribute("data-comment-id");
    const action = button.getAttribute("data-action");

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
        credentials: 'include'
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

document.querySelectorAll(".like-btn, .dislike-btn").forEach(button => {
    button.addEventListener("click", handleReaction);
});

document.querySelectorAll(".comment-like-btn, .comment-dislike-btn").forEach(button => {
    button.addEventListener("click", handleCommentReaction);
});

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
    const commentElement = document.getElementById(`comment-${commentId}`);
    if (!commentElement) return;

    // Create confirmation UI
    const confirmDiv = document.createElement('div');
    confirmDiv.className = 'alert alert-warning';
    confirmDiv.style.marginTop = '10px';
    
    const confirmText = document.createElement('span');
    confirmText.textContent = 'Are you sure you want to delete this comment? ';
    confirmDiv.appendChild(confirmText);
    
    const confirmBtn = document.createElement('button');
    confirmBtn.className = 'btn btn-sm btn-danger';
    confirmBtn.textContent = 'Yes';
    confirmBtn.style.marginRight = '10px';
    
    const cancelBtn = document.createElement('button');
    cancelBtn.className = 'btn btn-sm';
    cancelBtn.textContent = 'No';
    
    confirmDiv.appendChild(confirmBtn);
    confirmDiv.appendChild(cancelBtn);
    
    commentElement.appendChild(confirmDiv);
    
    cancelBtn.onclick = () => {
        confirmDiv.remove();
    };
    confirmBtn.onclick = () => {
        confirmDiv.remove();
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
                const messageElement = document.createElement('div');
                messageElement.className = 'alert alert-success';
                messageElement.textContent = 'Comment deleted successfully';
                
                commentElement.parentNode.insertBefore(messageElement, commentElement.nextSibling);
                commentElement.remove();
                
                setTimeout(() => {
                    messageElement.remove();
                }, 3000);
            } else {
                throw new Error('Failed to delete comment');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            const messageElement = document.createElement('div');
            messageElement.className = 'alert alert-error';
            messageElement.textContent = error.message;
            
            commentElement.parentNode.insertBefore(messageElement, commentElement.nextSibling);
            
            setTimeout(() => {
                messageElement.remove();
            }, 500);
        });
    };
}