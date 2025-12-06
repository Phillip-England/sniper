/// <reference lib="dom" />

import { SniperCore } from "./SniperCore";
/**
 * SECTION 5: INITIALIZATION
 */
document.addEventListener("DOMContentLoaded", async () => {
  const app = await SniperCore.new();
});
