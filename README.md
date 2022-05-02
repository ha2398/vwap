# vwap

Real-time VWAP (volume-weighted average price) calculation engine.

## Dependencies

To use this repository, the `make` utility is required. In addition, if you'd like to build the project and execute locally on your machine, Go v1.17+ is required. However, building and executing can also be done through Docker.

## Build

To build the project locally, run:

```
make install
```

To build the project using Docker, run:

```
make docker/build
```

## Usage

The `vwap` calculation engine logs in real time the VWAP values for each trading pair of interest, in a window of configurable size. To start the engine, run the following command:

<table>
<tr>
<th>Local</th>
<th>Docker</th>
</tr>
<tr>
<td>

```bash
make run
```

</td>
<td>

```bash
make docker/run
```

</td>
</tr>
</table>

For both cases, the following environment variables can be passed to customize the engine:

- **FEED_ENDPOINT**: WebSocket endpoint to read trading pair match data from, _e.g._, `wss://endpoint.company.com`.
- **TRADING_PAIRS**: Comma-separated list of trading pairs of interest to calculate VWAP for, _e.g._, `BTC-USD,ETH-BTC`.
- **WINDOW_SIZE**: Size of the sliding window to use when calculating VWAP. This has to be at least `1`.

## Design

This section presents design choices and implementation details for the project.

### Object model and data flow

All WebSocket messages are read as JSON objects. These objects are simple `map[string]interface{}`, so that any message can be read using the same underlying type. Since, for the VWAP calculation, we are only interested in the `match` or `last_match` message types, these maps are parsed to the `Match` structure when we observe these messages.

In order to allow for increased throughput of incoming WebSocket messages, one `goroutine` is spawned for reading messages, and another one is spawned for handling them. This way, the reader `goroutine` reads messages and place them in a buffered channel. The handler `goroutine` then feeds from this channel to handle new messages.

### Calculation Algorithm

The VWAP of a product over a window of `n` data points is defined as $\frac{\sum_{i=1}^{n} P_i * Q_i}{\sum_{i=1}^{n}Q_i} $, where `P_i` and `Q_i` are the price and quantity of a given data point `i`. Since we would like to calculate the VWAP for a sliding window, for every new data point, we also want to minimize the cost of performing this calculation, as it will be executed repeatedly.

By keeping track of the numerator and denominator for the fraction that calculates the VWAP, we can incorporate a new data point `j` into the calculation by adding its contribution to both the numerator and to the denominator. In addition, if the sliding window has reached its limit size, it will discard its oldest entry, meaning that we have to subtract its contribution from the numerator and from the denominator. Hence, we have that $VWAP_j = \frac{{\sum P*Q} -{P*Q}_{old} + {P*Q}_{new}}{{\sum Q} - Q_{old} + Q_{new}} $.

This calculation method has been implemented using a channel, which behaves as a queue for the sliding window, allowing us to easily pop the oldest data point and push the new one.

### Tests

Unit and integration tests have been added for the project. To run them, execute the following command:

```
make test
```