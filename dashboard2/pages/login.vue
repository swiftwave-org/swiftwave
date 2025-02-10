<template>
  <div class="flex h-full w-full flex-row">
    <!--  Content  -->
    <div
      class="relative flex h-full min-w-[60vw] select-none flex-col items-center bg-[#F9F8F8] pt-52"
    >
      <!--   Logo with title/subtitle   -->
      <div class="flex w-fit flex-row items-center justify-center gap-2">
        <!-- <img src="@/assets/images/logo.png" class="w-14" alt="swiftwave logo" /> -->
        <div class="flex flex-col items-start justify-between">
          <p class="font-prompt text-3xl">swiftwave</p>
          <p class="font-prompt text-base">open source paas</p>
        </div>
      </div>
      <!--    Heading  -->
      <p class="mt-32 font-comfortaa text-5xl">
        <span class="text-primary-600">Simple Lightweight</span>&nbsp;PaaS
      </p>
      <p class="mt-6 font-comfortaa text-5xl">for self-hosting</p>
      <!--   Button panel   -->
      <div
        class="absolute bottom-0 left-0 right-0 flex flex-row flex-wrap items-center justify-center gap-3 pb-6"
      >
        <!--        <p class="w-full text-center">Hemlo bro</p>-->
        <a
          class="action-btn"
          target="_blank"
          href="https://github.com/swiftwave-org/swiftwave"
        >
          Github
        </a>
        <a
          class="action-btn"
          target="_blank"
          href="https://github.com/swiftwave-org/swiftwave/issues/new/choose"
        >
          Report Bug
        </a>
        <a
          class="action-btn"
          target="_blank"
          href="https://slack.swiftwave.org/"
        >
          Join our community
        </a>
        <a
          class="action-btn"
          target="_blank"
          href="mailto:support@swiftwave.org"
        >
          Reach out to team
        </a>
        <a
          class="action-btn"
          target="_blank"
          href="https://swiftwave.org/docs/support_us/"
        >
          Support <b>Swiftwave</b>
        </a>
      </div>
    </div>
    <!--   Login form -->
    <div
      class="flex h-full w-full flex-col items-center justify-center px-6 py-12 lg:px-8"
    >
      <p class="w-fit text-5xl text-primary-600"></p>
      <div class="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
        <!-- Alert  -->
        <div
          v-if="authenticationStatus.visible"
          :class="{
            'border-red-500 bg-red-50': !authenticationStatus.success,
            'border-green-500 bg-green-50': authenticationStatus.success,
          }"
          class="mb-5 rounded border-s-4 p-4"
          role="alert"
        >
          <strong
            :class="{
              'text-red-800': !authenticationStatus.success,
              'text-green-800': authenticationStatus.success,
            }"
            class="block font-medium"
            >{{ authenticationStatus.message }}</strong
          >
        </div>

        <!--   Login Form   -->
        <!-- <form class="space-y-4" @keydown.enter.prevent="login"> -->
        <div>
          <label
            class="block text-sm font-medium leading-6 text-gray-900"
            for="username"
            >Username</label
          >
          <div class="mt-1">
            <input
              id="username"
              v-model="username"
              autocomplete="username"
              class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
              name="username"
              placeholder="Enter username"
              required
              type="text"
            />
          </div>
        </div>
        <div>
          <label
            class="block text-sm font-medium leading-6 text-gray-900"
            for="password"
            >Password</label
          >
          <div class="mt-1">
            <input
              id="password"
              v-model="password"
              autocomplete="current-password"
              class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
              placeholder="Enter password"
              required
              type="password"
            />
          </div>
        </div>
        <div v-if="authenticationStatus.totp_required">
          <label
            class="block text-sm font-medium leading-6 text-gray-900"
            for="2fa_code"
            >Provide 2FA Code</label
          >
          <div class="mt-2">
            <v-otp-input
              :num-inputs="6"
              input-classes="otp-input"
              :style="{ justifyContent: 'space-between' }"
              :placeholder="['*', '*', '*', '*', '*', '*']"
              v-model:value="totp"
              @on-change="(v) => (totp = v)"
            />
          </div>
        </div>
        <div class="py-2">
          <button @click="login" class="w-full">Sign in</button>
        </div>
        <!-- </form> -->
      </div>
    </div>
  </div>
</template>

<script setup>
const username = ref("serverabex");
const password = ref("Abex@10000Work");
const totp = ref("");
const authenticationStatus = reactive({
  visible: false,
  success: false,
  message: "",
  totp_required: false,
});
const authStore = useAuthStore();
const router = useRouter();

const login = async (e) => {
  e.preventDefault();
  console.log("Login button clicked");

  let res = await authStore.login(username.value, password.value, totp.value);
  if (res.totp_required) {
    authenticationStatus.totp_required = res.totp_required;
  } else {
    authenticationStatus.success = res.success;
    authenticationStatus.message = res.message;
    authenticationStatus.visible = true;
    authenticationStatus.totp_required =
      authenticationStatus.totp_required || res.totp_required;
    if (res.success) {
      // check if `redirect` is in the query
      if (router.currentRoute.value.query.redirect) {
        await router.push(router.currentRoute.value.query.redirect);
        return;
      }
      window.open(router.resolve({ name: "Applications" }).href, "_self");
    }
  }
};
</script>

<style lang="scss" scoped></style>
