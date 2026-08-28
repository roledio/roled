export function decodeJwtPayload(token: string): object | null {
  try {
    // Split the token into its parts
    const parts = token.split('.');
    if (parts.length !== 3) {
      throw new Error('Invalid JWT format');
    }

    // The payload is the second part (index 1)
    const base64UrlPayload = parts[1];

    // Base64Url decode the payload (handle URL-safe characters)
    const base64 = base64UrlPayload.replace(/-/g, '+').replace(/_/g, '/');

    // Decode the Base64 string to a raw JSON string
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    // Parse the JSON string into a JavaScript object
    return JSON.parse(jsonPayload);
  } catch (error) {
    console.error('Error decoding JWT payload:', error);
    return null;
  }
}
