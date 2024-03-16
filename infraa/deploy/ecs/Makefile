run:
	AWS_ACCOUNT=$(AWS_ACCOUNT) ../../build/terrawrapper.sh validate
	AWS_ACCOUNT=$(AWS_ACCOUNT) ../../build/terrawrapper.sh plan -out ./tfstates/dashsotdash.tfplan
	AWS_ACCOUNT=$(AWS_ACCOUNT) ../../build/terrawrapper.sh apply "./tfstates/dashsotdash.tfplan"
	chmod 600 dash-tf-key-pair.pem

destroy:
	AWS_ACCOUNT=$(AWS_ACCOUNT) ../../build/terrawrapper.sh destroy 

output:
	AWS_ACCOUNT=$(AWS_ACCOUNT) ../../build/terrawrapper.sh output

destroy.redis:
	AWS_ACCOUNT=$(AWS_ACCOUNT) ../../build/terrawrapper.sh destroy -target=aws_elasticache_cluster.redis_cache_cluster
