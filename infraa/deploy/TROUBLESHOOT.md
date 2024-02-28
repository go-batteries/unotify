## Troubleshooting

```shell
make run
```

This is to deploy aws instance using terraform scripts.


```shell
cat /var/log/cloud-init-output.log
```

This shows the output of instance bootstraping


```shell
TOKEN=`curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600"` \
&& curl -H "X-aws-ec2-metadata-token: $TOKEN" -v http://169.254.169.254/latest/user-data/
```

This to check the passed user-data
