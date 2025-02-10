import { defineStore } from "pinia";
import { computed, ref } from "vue";
import axios from "axios";
import { jwtDecode } from "jwt-decode";
import moment from "moment";

export const useAuthStore = defineStore("auth_details", () => {
  const isLoggedIn = ref(false);
  const accessToken = ref("");
  const isLoggingInProgress = ref(false);
  const currentUsername = ref("");
  const currentTime = ref(Date.now());

  function fetchBearerToken() {
    console.log("[AUTH] Fetching bearer token");
    if (isLoggedIn.value) {
      console.log("[AUTH] User is logged in, returning token");
      return "Bearer " + accessToken.value;
    }
    console.log("[AUTH] User is not logged in, returning empty token");
    return "";
  }

  function setCredential(token: string) {
    console.log("[AUTH] Setting credentials", token);
    accessToken.value = token;

    console.log("[AUTH] Storing token in localStorage");
    localStorage.setItem("token", token);

    isLoggedIn.value = true;
    isLoggingInProgress.value = true;
    setTimeout(() => {
      console.log("[AUTH] Login progress completed");
      isLoggingInProgress.value = false;
    }, 1000);
  }

  async function login(username: string, password: string, totp: string) {
    console.log("[AUTH] Attempting login for user:", username);
    const formData = new FormData();
    formData.append("username", username);
    formData.append("password", password);
    formData.append("totp", totp);

    const config = {
      method: "post",
      url: `${useRuntimeConfig().public.apiBaseUrl}/auth/login`,
      data: formData,
    };

    try {
      console.log("[AUTH] Sending login request");
      const res = await axios.request(config);
      const resData = res.data;
      console.log("[AUTH] Login successful", resData);
      setCredential(resData.token);
      return {
        success: true,
        message: "Logged in successfully!",
        totp_required: false,
      };
    } catch (e: any) {
      console.log("[AUTH] Login failed:", e.message);
      if (e.response) {
        return {
          success: false,
          message: e.response.data.message || "Unexpected error",
          totp_required: e.response.data.totp_required,
        };
      }
      return {
        success: false,
        message: "Failed to send request",
        totp_required: false,
      };
    }
  }

  function logout() {
    console.log("[AUTH] Logging out user");
    isLoggedIn.value = false;
    if (process.client) {
      console.log("[AUTH] Clearing localStorage");
      localStorage.clear();
    }
    isLoggingInProgress.value = true;
    setTimeout(() => {
      console.log("[AUTH] Redirecting to login page");
      navigateTo("/login");
    }, 500);
  }

  async function checkAuthStatus() {
    console.log("[AUTH] Checking authentication status");
    try {
      const token = localStorage.getItem("token");
      if (token) {
        console.log("[AUTH] Token found, verifying with server");
        const config = {
          method: "get",
          url: `${useRuntimeConfig().public.apiBaseUrl}/verify-auth`,
          headers: {
            Authorization: `Bearer ${token}`,
          },
        };
        const res = await axios.request(config);
        console.log("[AUTH] Token verification response:", res.status);
        return res.status === 200;
      }
    } catch (e: any) {
      console.log("[AUTH] Token verification failed:", e.message);
      if (e.isAxiosError) {
        if (e.message.includes("Network Error")) {
          return true;
        }
      }
      return false;
    }
    return false;
  }

  async function logoutOnInvalidToken(callback: () => void) {
    console.log("[AUTH] Checking token validity");
    if (!isLoggedIn.value) return;
    const isTokenValid = await checkAuthStatus();
    if (!isTokenValid) {
      console.log("[AUTH] Token invalid, executing logout callback");
      callback();
    }
  }

  function startAuthChecker(callback: () => void) {
    console.log("[AUTH] Starting auth checker");
    if (process.client) {
      setInterval(() => logoutOnInvalidToken(callback), 5000);
    }
  }

  const sessionRelativeTimeoutStatus = computed(() => {
    if (!process.client || !isLoggedIn.value) return "";

    try {
      console.log("[AUTH] Calculating session timeout status");
      const token = localStorage.getItem("token");
      if (token) {
        const decoded = jwtDecode(token);
        currentUsername.value = (decoded as any).username ?? "";
        const exp = moment(new Date((decoded as any).exp * 1000));
        return moment.duration(exp.diff(currentTime.value)).humanize(true);
      }
    } catch (e) {
      console.log("[AUTH] Error calculating session timeout:", e);
      return "N/A";
    }
    return "";
  });

  async function fetchSWVersion() {
    console.log("[AUTH] Fetching software version");
    if (!isLoggedIn.value) return "...";
    try {
      const config = {
        method: "get",
        url: `${useRuntimeConfig().public.apiBaseUrl}/version`,
        headers: {
          Authorization: fetchBearerToken(),
        },
      };
      const res = await axios.request(config);
      console.log("[AUTH] Software version fetched successfully");
      return res.data;
    } catch (e) {
      console.log("[AUTH] Error fetching software version:", e);
      return "N/A";
    }
  }

  function initializeAuth() {
    console.log("[AUTH] Initializing authentication");
    if (process.client) {
      const token = localStorage.getItem("token");
      if (token) {
        console.log("[AUTH] Found existing token, setting credentials");
        setCredential(token);
      }

      setInterval(() => {
        currentTime.value = Date.now();
      }, 10000);
    }
  }

  return {
    isLoggedIn,
    isLoggingInProgress,
    fetchBearerToken,
    login,
    logout,
    setCredential,
    startAuthChecker,
    sessionRelativeTimeoutStatus,
    fetchSWVersion,
    currentUsername,
    initializeAuth,
  };
});
