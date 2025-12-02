/**
 * SECTION 2: AUDIO MANAGER CLASS
 */
export type SoundType = 'sniper-on' | 'sniper-off' | 'sniper-exit' | 'click' | 'sniper-clear' | 'sniper-copy' | 'sniper-search' | 'sniper-visit';

export class AudioManager {
  private sounds: Record<SoundType, HTMLAudioElement>;

  constructor() {
    this.sounds = {
      'sniper-on': new Audio('/static/on.wav'),
      'sniper-off': new Audio('/static/off.wav'),
      'sniper-exit': new Audio('/static/exit.wav'),
      'sniper-clear': new Audio('/static/clear.wav'),
      'sniper-copy': new Audio('/static/copy.wav'),
      'sniper-search': new Audio('/static/search.wav'), 
      'sniper-visit': new Audio('/static/visit.wav'), 
      'click': new Audio('/static/click.wav'),
    };
    this.preloadSounds();
  }

  private preloadSounds() {
    Object.values(this.sounds).forEach((sound) => {
      sound.preload = 'auto';
      sound.load();
    });
  }

  public play(type: SoundType) {
    const audio = this.sounds[type];
    audio.currentTime = 0;
    audio.play().catch((e) => {
      console.warn(`Audio playback failed for ${type}`, e);
    });
  }
}