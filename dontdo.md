### Don't do

- `addDeploymentLog` function is safe to use without transaction at the time of using transaction in another thread.
- Except that above one, don't run any function without transaction at the time of using transaction.
- Avoid to use `dbWithoutTx` until unless required. Use carefully as it can lead to database deadlock condition.
- Don't use just Read Lock while closing channel, do RW lock