## BC Ferries API - [bcferriesapi.ca](https://bcferriesapi.ca)

🛳 The only public API for retrieving current data on BC Ferries sailings.

## How It's Made

BC Ferries API is a Go api connected to a web scraper. The scraper is made with Goquerey. It runs every minute and saves the relevant, formatted data to a json file. When the api recieves a request, it checks the validity of the request, then returns the specified data.

The frontend is made with HTML, Bootstrap, and Javascript. When the user hits the request button, it runs a small script that calls the api with the user's request, and displays the info on the page.

## API Reference

This API uses the route codes used by BC Ferries, they are:

```
"TSA" -> Tsawwassen
"SWB" -> Swartz Bay
"SGI" -> Southern Gulf Islands
"DUK" -> Duke Point (Nanaimo)
"FUL" -> Fulford Harbour (Salt Spring Island)
"HSB" -> Horseshoe Bay
"NAN" -> Departure Bay (Nanaimo)
"LNG" -> Langford
"BOW" -> Bowen Island
```

The api runs on the format:

`https://www.bcferriesapi.ca/api/<departure-terminal>/<destination-terminal>`

You can be as specific as needed, the fewer arguments there are, the more data you will recieve.

For example, `https://www.bcferriesapi.ca/api/tsawwassen` will return all data for all sailings leaving from Tsawwassen.

A request to `https://www.bcferriesapi.ca/api/` will return a full schedule for all terminals.

Options for each are as follows:

### departure-terminal

```
"TSA"
"SWB"
"HSB"
"DUK"
"LNG"
"NAN"
```

### destination-terminal

Note: destination terminal must correspond with departure terminal, for example you can't put tsawwassen to langdale, it will return an error.

```
"TSA": [
    "SWB"
    "SGI"
    "DUK"
]
"SWB": [
    "TSA"
    "FUL"
    "SGI"
]
"HSB": [
    "NAN"
    "LNG"
    "BOW"
]
"DUK": [
    "TSA"
]
"LNG": [
    "HSB"
]
"NAN": [
    "HSB"
]
```

## Sample response

This is a sample response from `https://www.bcferriesapi.ca/api/`.

```
{
    "schedule": {
        "DUK": {
            "TSA": {
                "sailingDuration": "2h 0m",
                "sailings": [
                    {
                        "time": "12:45PM",
                        "fill": 88,
                        "CarFill": 88,
                        "oversizeFill": 89,
                        "vesselName": "Queen of Alberni",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:15PM",
                        "fill": 65,
                        "CarFill": 70,
                        "oversizeFill": 60,
                        "vesselName": "Coastal Inspiration",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:45PM",
                        "fill": 70,
                        "CarFill": 48,
                        "oversizeFill": 91,
                        "vesselName": "Queen of Alberni",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:15PM",
                        "fill": 48,
                        "CarFill": 36,
                        "oversizeFill": 60,
                        "vesselName": "Coastal Inspiration",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:45PM",
                        "fill": 45,
                        "CarFill": 15,
                        "oversizeFill": 74,
                        "vesselName": "Queen of Alberni",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:15AM",
                        "fill": 32,
                        "CarFill": 14,
                        "oversizeFill": 50,
                        "vesselName": "Coastal Inspiration",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:45AM",
                        "fill": 61,
                        "CarFill": 60,
                        "oversizeFill": 62,
                        "vesselName": "Queen of Alberni",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:15AM",
                        "fill": 56,
                        "CarFill": 50,
                        "oversizeFill": 61,
                        "vesselName": "Coastal Inspiration",
                        "vesselStatus": ""
                    }
                ]
            }
        },
        "HSB": {
            "BOW": {
                "sailingDuration": "0h 20m",
                "sailings": [
                    {
                        "time": "11:25AM",
                        "fill": 26,
                        "CarFill": 25,
                        "oversizeFill": 29,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "12:35PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "2:20PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:30PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "4:35PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:45PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "6:50PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:20PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:30PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:30PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:50AM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "6:50AM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:00AM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Capilano",
                        "vesselStatus": ""
                    }
                ]
            },
            "LNG": {
                "sailingDuration": "0h 40m",
                "sailings": [
                    {
                        "time": "11:55AM",
                        "fill": 67,
                        "CarFill": 57,
                        "oversizeFill": 91,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "2:10PM",
                        "fill": 33,
                        "CarFill": 24,
                        "oversizeFill": 56,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "4:20PM",
                        "fill": 32,
                        "CarFill": 33,
                        "oversizeFill": 27,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "6:35PM",
                        "fill": 27,
                        "CarFill": 24,
                        "oversizeFill": 34,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:45PM",
                        "fill": 19,
                        "CarFill": 16,
                        "oversizeFill": 26,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:55PM",
                        "fill": 5,
                        "CarFill": 16,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:30AM",
                        "fill": 43,
                        "CarFill": 28,
                        "oversizeFill": 80,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:45AM",
                        "fill": 46,
                        "CarFill": 34,
                        "oversizeFill": 76,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "11:55AM",
                        "fill": 45,
                        "CarFill": 34,
                        "oversizeFill": 71,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    }
                ]
            },
            "NAN": {
                "sailingDuration": "1h 40m",
                "sailings": [
                    {
                        "time": "1:00PM",
                        "fill": 48,
                        "CarFill": 56,
                        "oversizeFill": 26,
                        "vesselName": "Queen of Cowichan",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:45PM",
                        "fill": 48,
                        "CarFill": 51,
                        "oversizeFill": 40,
                        "vesselName": "Queen of Oak Bay",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:40PM",
                        "fill": 40,
                        "CarFill": 49,
                        "oversizeFill": 17,
                        "vesselName": "Queen of Cowichan",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:05PM",
                        "fill": 30,
                        "CarFill": 34,
                        "oversizeFill": 22,
                        "vesselName": "Queen of Oak Bay",
                        "vesselStatus": ""
                    },
                    {
                        "time": "6:15AM",
                        "fill": 34,
                        "CarFill": 31,
                        "oversizeFill": 42,
                        "vesselName": "Queen of Oak Bay",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:25AM",
                        "fill": 46,
                        "CarFill": 52,
                        "oversizeFill": 32,
                        "vesselName": "Queen of Cowichan",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:40AM",
                        "fill": 44,
                        "CarFill": 52,
                        "oversizeFill": 23,
                        "vesselName": "Queen of Oak Bay",
                        "vesselStatus": ""
                    }
                ]
            }
        },
        "LNG": {
            "HSB": {
                "sailingDuration": "0h 40m",
                "sailings": [
                    {
                        "time": "1:05PM",
                        "fill": 42,
                        "CarFill": 39,
                        "oversizeFill": 52,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:15PM",
                        "fill": 38,
                        "CarFill": 35,
                        "oversizeFill": 47,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:25PM",
                        "fill": 27,
                        "CarFill": 21,
                        "oversizeFill": 43,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:40PM",
                        "fill": 10,
                        "CarFill": 9,
                        "oversizeFill": 15,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:50PM",
                        "fill": 3,
                        "CarFill": 2,
                        "oversizeFill": 6,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "6:20AM",
                        "fill": 14,
                        "CarFill": 3,
                        "oversizeFill": 42,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:40AM",
                        "fill": 36,
                        "CarFill": 35,
                        "oversizeFill": 38,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:50AM",
                        "fill": 36,
                        "CarFill": 31,
                        "oversizeFill": 50,
                        "vesselName": "Queen of Coquitlam",
                        "vesselStatus": ""
                    }
                ]
            }
        },
        "NAN": {
            "HSB": {
                "sailingDuration": "1h 40m",
                "sailings": [
                    {
                        "time": "1:00PM",
                        "fill": 74,
                        "CarFill": 84,
                        "oversizeFill": 50,
                        "vesselName": "Queen of Oak Bay",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:20PM",
                        "fill": 69,
                        "CarFill": 77,
                        "oversizeFill": 48,
                        "vesselName": "Queen of Cowichan",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:55PM",
                        "fill": 68,
                        "CarFill": 72,
                        "oversizeFill": 58,
                        "vesselName": "Queen of Oak Bay",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:10PM",
                        "fill": 21,
                        "CarFill": 16,
                        "oversizeFill": 35,
                        "vesselName": "Queen of Oak Bay",
                        "vesselStatus": ""
                    },
                    {
                        "time": "6:15AM",
                        "fill": 25,
                        "CarFill": 26,
                        "oversizeFill": 22,
                        "vesselName": "Queen of Cowichan",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:25AM",
                        "fill": 43,
                        "CarFill": 57,
                        "oversizeFill": 5,
                        "vesselName": "Queen of Oak Bay",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:40AM",
                        "fill": 63,
                        "CarFill": 77,
                        "oversizeFill": 26,
                        "vesselName": "Queen of Cowichan",
                        "vesselStatus": ""
                    }
                ]
            }
        },
        "SWB": {
            "FUL": {
                "sailingDuration": "0h 35m",
                "sailings": [
                    {
                        "time": "11:00AM",
                        "fill": 67,
                        "CarFill": 67,
                        "oversizeFill": 0,
                        "vesselName": "Skeena Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "1:00PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Skeena Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:00PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Skeena Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:00PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Skeena Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:00PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Skeena Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:00PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Skeena Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:00AM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Skeena Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:00AM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Skeena Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "11:00AM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Skeena Queen",
                        "vesselStatus": ""
                    }
                ]
            },
            "SGI": {
                "sailingDuration": "Varies",
                "sailings": [
                    {
                        "time": "2:20PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Cumberland",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:10PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Mayne Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "4:20PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Cumberland",
                        "vesselStatus": ""
                    },
                    {
                        "time": "6:40PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Cumberland",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:55PM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Mayne Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:00AM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Mayne Queen",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:05AM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Cumberland",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:15AM",
                        "fill": 0,
                        "CarFill": 0,
                        "oversizeFill": 0,
                        "vesselName": "Queen of Cumberland",
                        "vesselStatus": ""
                    }
                ]
            },
            "TSA": {
                "sailingDuration": "1h 35m",
                "sailings": [
                    {
                        "time": "11:00AM",
                        "fill": 100,
                        "CarFill": 100,
                        "oversizeFill": 100,
                        "vesselName": "Spirit of Vancouver Island",
                        "vesselStatus": ""
                    },
                    {
                        "time": "1:00PM",
                        "fill": 100,
                        "CarFill": 100,
                        "oversizeFill": 100,
                        "vesselName": "Spirit of British Columbia",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:00PM",
                        "fill": 75,
                        "CarFill": 93,
                        "oversizeFill": 58,
                        "vesselName": "Spirit of Vancouver Island",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:00PM",
                        "fill": 79,
                        "CarFill": 93,
                        "oversizeFill": 65,
                        "vesselName": "Spirit of British Columbia",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:00PM",
                        "fill": 73,
                        "CarFill": 95,
                        "oversizeFill": 52,
                        "vesselName": "Spirit of Vancouver Island",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:00PM",
                        "fill": 67,
                        "CarFill": 83,
                        "oversizeFill": 52,
                        "vesselName": "Spirit of British Columbia",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:00AM",
                        "fill": 85,
                        "CarFill": 87,
                        "oversizeFill": 83,
                        "vesselName": "Spirit of Vancouver Island",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:00AM",
                        "fill": 73,
                        "CarFill": 95,
                        "oversizeFill": 51,
                        "vesselName": "Spirit of British Columbia",
                        "vesselStatus": ""
                    },
                    {
                        "time": "11:00AM",
                        "fill": 79,
                        "CarFill": 89,
                        "oversizeFill": 68,
                        "vesselName": "Spirit of Vancouver Island",
                        "vesselStatus": ""
                    }
                ]
            }
        },
        "TSA": {
            "DUK": {
                "sailingDuration": "2h 0m",
                "sailings": [
                    {
                        "time": "12:45PM",
                        "fill": 68,
                        "CarFill": 68,
                        "oversizeFill": 68,
                        "vesselName": "Coastal Inspiration",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:15PM",
                        "fill": 70,
                        "CarFill": 54,
                        "oversizeFill": 85,
                        "vesselName": "Queen of Alberni",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:45PM",
                        "fill": 56,
                        "CarFill": 31,
                        "oversizeFill": 81,
                        "vesselName": "Coastal Inspiration",
                        "vesselStatus": ""
                    },
                    {
                        "time": "8:15PM",
                        "fill": 58,
                        "CarFill": 29,
                        "oversizeFill": 86,
                        "vesselName": "Queen of Alberni",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:45PM",
                        "fill": 45,
                        "CarFill": 4,
                        "oversizeFill": 86,
                        "vesselName": "Coastal Inspiration",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:15AM",
                        "fill": 48,
                        "CarFill": 13,
                        "oversizeFill": 83,
                        "vesselName": "Queen of Alberni",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:45AM",
                        "fill": 54,
                        "CarFill": 35,
                        "oversizeFill": 73,
                        "vesselName": "Coastal Inspiration",
                        "vesselStatus": ""
                    },
                    {
                        "time": "10:15AM",
                        "fill": 74,
                        "CarFill": 80,
                        "oversizeFill": 68,
                        "vesselName": "Queen of Alberni",
                        "vesselStatus": ""
                    }
                ]
            },
            "SGI": {
                "sailingDuration": "Varies",
                "sailings": [
                    {
                        "time": "7:10PM",
                        "fill": 100,
                        "CarFill": 100,
                        "oversizeFill": 100,
                        "vesselName": "Salish Eagle",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:55AM",
                        "fill": 100,
                        "CarFill": 100,
                        "oversizeFill": 100,
                        "vesselName": "Salish Eagle",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:10PM",
                        "fill": 98,
                        "CarFill": 100,
                        "oversizeFill": 97,
                        "vesselName": "Salish Eagle",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:55AM\n\n\n(Mar25,2022)",
                        "fill": 98,
                        "CarFill": 100,
                        "oversizeFill": 97,
                        "vesselName": "Salish Eagle",
                        "vesselStatus": ""
                    }
                ]
            },
            "SWB": {
                "sailingDuration": "1h 35m",
                "sailings": [
                    {
                        "time": "11:00AM",
                        "fill": 100,
                        "CarFill": 100,
                        "oversizeFill": 100,
                        "vesselName": "Spirit of British Columbia",
                        "vesselStatus": ""
                    },
                    {
                        "time": "1:00PM",
                        "fill": 100,
                        "CarFill": 100,
                        "oversizeFill": 100,
                        "vesselName": "Spirit of Vancouver Island",
                        "vesselStatus": ""
                    },
                    {
                        "time": "3:00PM",
                        "fill": 88,
                        "CarFill": 100,
                        "oversizeFill": 76,
                        "vesselName": "Spirit of British Columbia",
                        "vesselStatus": ""
                    },
                    {
                        "time": "5:00PM",
                        "fill": 80,
                        "CarFill": 93,
                        "oversizeFill": 67,
                        "vesselName": "Spirit of Vancouver Island",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:00PM",
                        "fill": 67,
                        "CarFill": 83,
                        "oversizeFill": 50,
                        "vesselName": "Spirit of British Columbia",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:00PM",
                        "fill": 62,
                        "CarFill": 51,
                        "oversizeFill": 74,
                        "vesselName": "Spirit of Vancouver Island",
                        "vesselStatus": ""
                    },
                    {
                        "time": "7:00AM",
                        "fill": 88,
                        "CarFill": 86,
                        "oversizeFill": 90,
                        "vesselName": "Spirit of British Columbia",
                        "vesselStatus": ""
                    },
                    {
                        "time": "9:00AM",
                        "fill": 76,
                        "CarFill": 99,
                        "oversizeFill": 53,
                        "vesselName": "Spirit of Vancouver Island",
                        "vesselStatus": ""
                    },
                    {
                        "time": "11:00AM",
                        "fill": 63,
                        "CarFill": 62,
                        "oversizeFill": 64,
                        "vesselName": "Spirit of British Columbia",
                        "vesselStatus": ""
                    }
                ]
            }
        }
    },
    "scrapedAt": "2022-03-23T10:57:52.096177-07:00"
}
```
