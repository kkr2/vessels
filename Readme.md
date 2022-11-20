# Vessels

Vessels is a demo project 

## How to run

First step is to have [docker](https://www.docker.com/products/docker-desktop/) installed to your local machine.

If this is your first run we need to build the containers.
```bash
make docker_build
```
If you want to run it 

```bash
make docker_run
```
## Usage

### POST `/api/v1/vessels`

#### Request (as provided in samples)
```json
{
    "imo": 345678,
    "draught" : 10.2,
    "routes": [
        [
            {
              "date": "2022-03-02T21:55:00Z",
              "longitude": -81.1,
              "latitude": 32.08333300000001
            },
            {
              "date": "2022-03-02T22:03:00Z",
              "longitude": -81.0831946,
              "latitude": 32.0808137
            },
       
        ],
        [
            {
              "date": "2022-03-02T21:55:00Z",
              "longitude": -81.1,
              "latitude": 32.08333300000001
            },
            {
              "date": "2022-03-02T22:03:00Z",
              "longitude": -81.0831946,
              "latitude": 32.0808137
            },

        ]
    ]
}
```
#### Response
```json
[
    {
        "ConsumtionInMetricTons": 62.94952536010629,
        "ConsumptionInCO2": 196.02482197137098
    },
    {
        "ConsumtionInMetricTons": 54.443613358700766,
        "ConsumptionInCO2": 169.53741199899417
    }
]
```

## CSV cleaning
CSV's provided were modified to have the same data model. 
On `model2.csv` only the raws with `added_resistance` 0 are taken into consideration. Also `imo` was not the same and was converted to 123456 for all the file.
Other csv files required column renaming and column removal.

All clean csv files are on `/csv` folder later to be mounted to postgres container and imported via migration

## How it works (General strategy)

1) For a given vessel `imo` we find the closest `draught` and retrieve all records that match with closest `draught`. This part is important for all further staps since this records can be reused since all routes requred to be calculated have the same `draught`. (Saves a lot of DB requests)

2) We create a `PointToPoint` data structure that represents the distance between 2 data points given by the request. On the next steps we populate this `PointToPoint` data structure with information like weather, distance and fuel consumption.

3) We populate `PointToPoint` with avg speed based on distance and time needed for the vessel to float from 1st to 2nd location.This helps us make a more accurate fuel consumtion calculation on next steps.

4) We populate `PointToPoint` with weather information retrieved by an external endoint provided to us. This endpoint recieves a specific day end returns the `beaufort` (avg wind level for that day). This also helps us make a more accurate fuel consumtion calculation. This client has added cache so it helps with performance.

5) In this step we add `avgFuelConsumtion` to every `PointToPoint` we have. We do this by using data we got from step 1 that guarentees us that this fueldata is the closest with the provided `draught`. By providing metadata gathered from previous step, we sort this data based on smallest delta on `weather` and if the delta is the same we sort based on `speed`. This provides us the closest avg fuel consumtion for every `PointToPoint` based on hiarchy included in project description `draught` > `weather` > `speed`

6) Based on `timeDuration` for vessel to float from a location to another and also the `avgFuelConsumtion` we are able to calculate `exactFuelConsumtion`. This means we have an exact fuel consumation in metric tons for the vessel to float from pont 1 to point 2.

7) We add all `exactFuelConsumtion` from every `PointToPoint` we have and this returns a pretty accurate fuel consumtion per `Route`

## What could be better

### Tests
Missing due to time restriction. Would like to make some for the domain, and usecase where most of the calculations happen.

### Documentation
Due to the nature of the project I would have liked to leave more comments throughout the part where calculations are made
### Weather Calculation
Function that calculates avg beaufort between 2 provided data points has a flaw that if data provided is more that 2 days appart the calculation is inacurate. Also the calculation between 2 consecutive days should be more accurate but for simplicity it returns avg of the 2.
### Weather Cache
Should be swapped with a production ready cache (redis,memcached or in-memory with golang lib) that have TTL and eviction policy. 

### Eco flag
Not sure what I should have provided if the flag is enabled

### Endpoint
If the service had multiple entities the routing should have been more accurate like `/api/v1/vessels/{vesselId}/fuelconsumtion`

### Bonus 
For achiving a service that responds to 25k rps, I would think of an async api doing more or less the same that the syncronus api calculates.
Comand and Query should be seperate similar to (CQRS).

These are the components for 

1) Command api for submiting the request that puts the request to a queue and returns an ACK as quickly as possible.

2) Queue to store the request and also it serves as a backpressure. 

3) Service that calculates the fuel consumtion based on the requests coming from the queue.

4) DB that has fuel consumtion map could be scaled by adding slaves.

5) Query api so client can get results submited for calculation. We can also implement a bulk api for getting a lot of responses submited at the same time.



