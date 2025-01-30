document.addEventListener('DOMContentLoaded', function() {
    document.querySelectorAll('.action-btn').forEach(button => {
        button.addEventListener('click', function() {
            const postId = this.getAttribute('data-post-id');
            const action = this.getAttribute('data-action');
            handleReaction(postId, action);
        });
    });
});

function handleReaction(postId, action) {
    const sessionToken = document.cookie
        .split('; ')
        .find(row => row.startsWith('session_token='))
        ?.split('=')[1];

    console.log("Attempting reaction:", { postId, action, sessionToken });

    if (!sessionToken) {
        window.location.href = '/signin';
        return;
    }

    fetch('/react', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        },
        body: JSON.stringify({
            post_id: parseInt(postId),
            reaction_type: action
        }),
        credentials: 'include'
    })
    .then(async res => {
        if (!res.ok) {
            const text = await res.text();
            console.error('Response:', text);
            throw new Error(`HTTP error! status: ${res.status}`);
        }
        return res.json();
    })
    .then(data => {
        console.log("Reaction response:", data);
        if (data.error) {
            throw new Error(data.error);
        }
        document.querySelector(`#likes-${postId}`).textContent = data.likes;
        document.querySelector(`#dislikes-${postId}`).textContent = data.dislikes;
    })
    .catch(error => {
        console.error('Error:', error);
        alert('Failed to update reaction');
    });
}