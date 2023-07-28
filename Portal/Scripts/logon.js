var Logon={
    init:function(){

        



        const loginForm = document.getElementById("loginform");
        console.log("submit Login")
        loginForm.addEventListener("submit", async (event) => {
            event.preventDefault();

            // Get the form data
            const formData = new FormData(loginForm);
            const username = formData.get("username");
            const password = formData.get("password");
            /*
            // Send a POST request to the login API to validate the user credentials
            const response = await fetch("/api/login", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    username,
                    password
                })
            });

            if (response.ok) {
            // If the login API call is successful, store the generated token in the session and redirect to the home page
            const data = await response.json();
            sessionStorage.setItem("token", data.token);
            window.location.href = "/home";
            } else {
            // If the login API call fails, display an error message to the user
            const errorData = await response.json();
            alert(`Login failed: ${errorData.message}`);
            } */

            login();

      });
    }
}
