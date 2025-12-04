/// <reference lib="dom" />

import { UIManager } from "./UIManager";
import { AudioManager } from "./AudioManager";
import { SniperCore } from "./SniperCore";
/**
 * SECTION 5: INITIALIZATION
 */
document.addEventListener('DOMContentLoaded', () => {
  const audioManager = new AudioManager();
  const uiManager = new UIManager();
  const app = new SniperCore(audioManager, uiManager);
});
