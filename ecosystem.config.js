module.exports = {
  apps : [{
    name   : "server",
    script : "./out/server",
    args   : "",
    watch: ["./out/server"]
  },{
    name   : "worker",
    script : "./out/worker",
    args   : "",
    watch: ["./out/worker"]
  }]
}

