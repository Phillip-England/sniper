from fastapi import FastAPI
import pyautogui

app = FastAPI()

@app.get("/")
def read_root():
    # This moves the mouse relative to its current position
    # (-20 means 20 pixels to the left, 0 means no vertical movement)
    try:
        pyautogui.moveRel(-20, 0)
        action_status = "Mouse moved 20 pixels to the left."
    except Exception as e:
        # This catches errors if the server is headless (no display)
        action_status = f"Failed to move mouse: {str(e)}"

    return {"Hello": "World", "Action": action_status}