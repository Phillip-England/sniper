dev:
	go run main.go

tw:
	tailwindcss -i ./static/input.css -o ./static/output.css --watch;

bundle:
	bun build ./client/index.ts --outdir ./static --watch;

server:
	uvicorn main:app --reload;

kill:
	lsof -ti:8000 | xargs kill -9 || true