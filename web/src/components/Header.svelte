<script>
export let appVersion = "";
</script>

<header>
  <div class="header-left">
    <svg
      class="logo"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor" />
      <path
        d="M2 17L12 22L22 17"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        fill="none"
      />
      <path
        d="M2 12L12 17L22 12"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        fill="none"
      />
    </svg>
    <h1>CoreOps Dashboard</h1>
  </div>
  <div class="header-right">
    <span class="version-text">{appVersion}</span>
    <a
      href="https://github.com/SSomov/docker-dashboard"
      target="_blank"
      rel="noopener noreferrer"
      class="header-link"
      aria-label="GitHub Repository"
    >
      <svg
        viewBox="0 0 24 24"
        fill="currentColor"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
        />
      </svg>
    </a>
    <a
      href="https://hub.docker.com/r/avt0x/docker-dashboard"
      target="_blank"
      rel="noopener noreferrer"
      class="header-link"
      aria-label="Docker Hub"
    >
      <svg
        class="docker-hub-logo"
        viewBox="0 0 24 24"
        fill="white"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M23.6096 7.78281C23.0211 7.38668 21.4751 7.21752 20.3513 7.52051C20.291 6.40098 19.7132 5.45739 18.6572 4.63407L18.2664 4.37177L18.0059 4.7654C17.4937 5.54269 17.2778 6.5787 17.3542 7.52051C17.4145 8.10079 17.6162 8.75281 18.0059 9.22603C16.543 10.0747 15.1944 9.88197 9.22236 9.88197H0.00204151C-0.0247244 11.2306 0.191901 13.8244 1.84139 15.9361C2.02376 16.1695 2.22326 16.395 2.44024 16.6124C3.78139 17.9553 5.80775 18.9403 8.83765 18.9428C13.4603 18.9471 17.4206 16.4482 19.8299 10.4069C20.6225 10.4198 22.7152 10.549 23.7395 8.57009C23.7645 8.5369 24 8.04548 24 8.04548L23.6096 7.78317V7.78281ZM6.01938 6.5498H3.42665V9.14252H6.01938V6.5498ZM9.36868 6.5498H6.77596V9.14252H9.36868V6.5498ZM12.7183 6.5498H10.1256V9.14252H12.7183V6.5498ZM16.068 6.5498H13.4753V9.14252H16.068V6.5498ZM2.66971 6.5498H0.0769861V9.14252H2.66971V6.5498ZM6.01938 3.27508H3.42665V5.8678H6.01938V3.27508ZM9.36868 3.27508H6.77596V5.8678H9.36868V3.27508ZM12.7183 3.27508H10.1256V5.8678H12.7183V3.27508ZM12.7183 0H10.1256V2.59272H12.7183V0Z"
          fill="white"
        />
      </svg>
    </a>
  </div>
</header>

<style>
  header {
    background-color: #4caf50;
    color: white;
    padding: 0.4rem 1rem;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  header h1 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
  }

  .logo {
    width: 24px;
    height: 24px;
    flex-shrink: 0;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .header-link {
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    text-decoration: none;
    transition: opacity 0.2s;
    flex-shrink: 0;
  }

  .header-link:hover {
    opacity: 0.7;
  }

  .header-link svg {
    width: 24px;
    height: 24px;
  }

  .docker-hub-logo {
    width: 120px;
    height: 37px;
    display: block;
    margin: 0;
    padding: 0;
  }

  .header-link:has(.docker-hub-logo) {
    padding: 0;
    margin: 0;
    width: auto;
    height: auto;
  }

  .version-text {
    color: white;
    font-size: 0.9rem;
    font-weight: 500;
    white-space: nowrap;
    opacity: 0.9;
  }
</style>

