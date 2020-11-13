# MonitorWebsiteMetrics
For Inserting Metrics endpoint is
http://ipAddr:8080/v1/insert

For Querying Metrics endpoint is
http://ipAddr:8080/v1/query


For adding more dimentions Edit priority.json
add the dimention name with its priority

Example Default file
{
  "Priority": {
    "country": 0,
    "device": 1
  }
}

After adding dimention
{
  "Priority": {
    "country": 0,
    "device": 1,
    "os": 2
  }
}

