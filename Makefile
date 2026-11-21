.PHONY: up down seed gen bench bench-stdlib bench-pgxpool bench-ent bench-sqlc compare profile pprof clean

up:
	docker compose up -d

wait-db:
	until docker compose exec -T postgres pg_isready -U bench -d bench; do sleep 1; done

down:
	docker compose down -v

seed:
	go run ./seed

mkdirs:
	mkdir -p results profiles

gen:
	cd variants/sqlc && sqlc generate

bench: mkdirs bench-stdlib bench-pgxpool bench-ent bench-sqlc

bench-stdlib:
	go test ./variants/stdlib/... -run '^$$' -bench . -benchmem -count=10 | tee results/stdlib.txt

bench-pgxpool:
	go test ./variants/pgxpool/... -run '^$$' -bench . -benchmem -count=10 | tee results/pgxpool.txt

bench-ent:
	go test ./variants/ent/... -run '^$$' -bench . -benchmem -count=10 | tee results/ent.txt

bench-sqlc:
	go test ./variants/sqlc/... -run '^$$' -bench . -benchmem -count=10 | tee results/sqlc.txt

compare:
	benchstat results/stdlib.txt results/pgxpool.txt results/ent.txt results/sqlc.txt

profile:
	PPROF=1 go test ./variants/pgxpool/... -run '^$$' -bench . -benchmem \
	  -cpuprofile profiles/pgxpool-cpu.prof -memprofile profiles/pgxpool-mem.prof

pprof:
	go tool pprof -http=:8081 profiles/pgxpool-cpu.prof

clean:
	rm -f profiles/*.prof results/*.txt
	docker compose down -v
