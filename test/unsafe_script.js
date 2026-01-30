// Vulnerability 3: DOM-based XSS
function displayUserParams() {
    const params = new URLSearchParams(window.location.search);
    const name = params.get("name");
    
    // Dangerous: Direct assignment to innerHTML
    document.getElementById("welcome").innerHTML = "Hello, " + name;
}

// Vulnerability 4: Dangerous Eval
function calculate(input) {
    // Dangerous: Execution of arbitrary code
    return eval(input);
}

// Vulnerability 5: Hardcoded API Token
const config = {
    apiKey: "sk_live_TEST_KEY_FOR_SECURITY_SCANNER_DETECTION_123456"
};
