dev:
	air

tw:
	tailwindcss -i ./static/input.css -o ./static/output.css --watch;

bundle:
	bun build ./client/index.ts --outdir ./static --watch;

kill:
	lsof -ti:8000 | xargs kill -9 || true
