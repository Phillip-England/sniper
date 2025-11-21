// client/index.ts
var sounds = {
  on: new Audio("/static/on.wav"),
  off: new Audio("/static/off.wav"),
  exit: new Audio("/static/exit.wav")
};
Object.values(sounds).forEach((sound) => {
  sound.preload = "auto";
  sound.load();
});
var playSound = (type) => {
  const audio = sounds[type];
  audio.currentTime = 0;
  audio.play().catch((e) => {
    console.warn(`Audio playback failed for ${type}. Check file path /static/${type}.wav`, e);
  });
};
var btn = document.getElementById("record-button");
var transcriptEl = document.getElementById("transcript");
var interimEl = document.getElementById("interim");
var outputContainer = document.getElementById("output-container");
var placeholder = document.getElementById("placeholder");
var statusText = document.getElementById("status-text");
var copyBtn = outputContainer.querySelector("button");
var greenDot = document.getElementById("green-dot");
var isRecording = false;
var isLogging = true;
var shouldContinue = false;
var recognition = null;
var SpeechRecognitionCtor = window.SpeechRecognition || window.webkitSpeechRecognition;
var defaultClasses = ["bg-red-600", "hover:scale-105", "hover:bg-red-500"];
var recordingClasses = ["bg-red-700", "animate-pulse", "ring-4", "ring-red-900"];
function updateGreenDot() {
  if (isRecording && isLogging) {
    greenDot.classList.remove("opacity-0");
  } else {
    greenDot.classList.add("opacity-0");
  }
}
if (!SpeechRecognitionCtor) {
  alert("Your browser does not support speech recognition. Try Chrome or Safari.");
} else {
  recognition = new SpeechRecognitionCtor;
  recognition.continuous = true;
  recognition.interimResults = true;
  recognition.lang = "en-US";
  recognition.onstart = () => {
    playSound("on");
    isRecording = true;
    updateGreenDot();
    btn.classList.remove(...defaultClasses);
    btn.classList.add(...recordingClasses);
    statusText.classList.remove("opacity-0");
    outputContainer.classList.remove("opacity-0", "translate-y-10");
    placeholder.textContent = "Listening...";
  };
  recognition.onend = () => {
    if (shouldContinue) {
      recognition?.start();
      return;
    }
    isRecording = false;
    updateGreenDot();
    btn.classList.remove(...recordingClasses);
    btn.classList.add(...defaultClasses);
    statusText.classList.add("opacity-0");
    placeholder.textContent = "Tap button to speak...";
    if (transcriptEl.innerText.trim().length > 0) {
      placeholder.classList.add("hidden");
    } else {
      placeholder.classList.remove("hidden");
    }
  };
  recognition.onresult = (event) => {
    let final = "";
    let interim = "";
    for (let i = event.resultIndex;i < event.results.length; ++i) {
      const result = event.results[i];
      if (!result)
        continue;
      const inner = result[0];
      if (!inner)
        continue;
      if (result.isFinal) {
        final += inner.transcript;
      } else {
        interim += inner.transcript;
      }
    }
    if (final) {
      const command = final.toLowerCase().trim().replace(/[.,?!]/g, "");
      if (command === "exit") {
        playSound("exit");
        transcriptEl.innerText = "";
        interimEl.innerText = "";
        placeholder.classList.remove("hidden");
        shouldContinue = false;
        stopRecording();
        return;
      }
      if (command === "stop") {
        playSound("off");
        transcriptEl.innerText = "";
        interimEl.innerText = "";
        placeholder.classList.remove("hidden");
        isLogging = false;
        updateGreenDot();
        return;
      }
      if (command === "start") {
        playSound("on");
        isLogging = true;
        updateGreenDot();
        return;
      }
      if (command === "clear") {
        transcriptEl.innerText = "";
        interimEl.innerText = "";
        placeholder.classList.remove("hidden");
        return;
      }
      if (isLogging) {
        transcriptEl.innerText += final + " ";
      }
    }
    if (isLogging) {
      interimEl.innerText = interim;
    } else {
      interimEl.innerText = "";
    }
    if (transcriptEl.innerText || interimEl.innerText) {
      placeholder.classList.add("hidden");
    } else {
      placeholder.classList.remove("hidden");
    }
  };
  recognition.onerror = (event) => {
    if (event.error === "no-speech")
      return;
    if (event.error === "not-allowed" || event.error === "service-not-allowed") {
      shouldContinue = false;
      stopRecording();
    }
  };
}
function startRecording() {
  shouldContinue = true;
  isLogging = true;
  if (recognition)
    recognition.start();
}
function stopRecording() {
  shouldContinue = false;
  if (recognition)
    recognition.stop();
}
if (btn) {
  btn.addEventListener("click", () => {
    if (isRecording) {
      stopRecording();
    } else {
      startRecording();
    }
  });
}
if (copyBtn) {
  copyBtn.onclick = null;
  copyBtn.addEventListener("click", () => {
    const text = transcriptEl.innerText;
    if (text) {
      navigator.clipboard.writeText(text);
      const originalText = copyBtn.innerText;
      copyBtn.innerText = "[ COPIED! ]";
      setTimeout(() => {
        copyBtn.innerText = originalText;
      }, 2000);
    }
  });
}
