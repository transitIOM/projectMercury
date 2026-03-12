![](https://github.com/user-attachments/assets/b6063d8a-04fc-409f-a50b-5003fd40f114)

Mercury is the API serving all schedule and realtime data for the transitIOM app, [Minervra](https://github.com/transitIOM/projectMinerva.git). It serves GTFS schedule ZIP files sourced from [Cura](https://github.com/transitIOM/projectCura), and computes and delievers realtime data based on live location data from [findmybus.im](https://findmybus.im).

## Installation (only reccomended for local development)
For local development clone the repository and install its dependencies. Ensure you have at least [Go 1.25](https://go.dev/doc/go1.25) installed. You also need to have a running instance of [Minio](https://www.min.io/), which can be deployed locally by running the [dev-docker-compose.yml](https://github.com/transitIOM/projectMercury/blob/main/dev-docker-compose.yml) file.

```bash
git clone https://github.com/transitIOM/projectMercury.git
cd projectMercury

go mod install
docker compose up -f dev-docker-compose.yml
go run main.go
```

This software is deployed on our own infrastructure. It is *NOT* reccomended to deploy it elsewhere without heavy modification for your own usecase.

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
Copyright 2026 transitIOM Ltd.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.