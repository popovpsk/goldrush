deploy:
	docker build -t goldrush .
	docker tag goldrush stor.highloadcup.ru/rally/dark_ferret
	docker push stor.highloadcup.ru/rally/dark_ferret