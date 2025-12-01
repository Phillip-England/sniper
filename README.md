# sniper-speech
Sniper speech is a system to control your computer using your voice. My passion for programming has caused me to develop pain in my left hand. This sent me on a journey to develop a system where i can continue to program without the use of my hands as much. Introducing sniper speech.

## Installation 
```bash 
go install github.com/phillip-england/sniper-speech@latest
```

## Wayland Display Errors
You may encounter issues when running `sniper` on system using wayland. For Ubuntu, I had to logout and switch my display settings on the login screen to X11 (xorg). It seems `robotgo` has issues interacting with the mouse when using wayland.