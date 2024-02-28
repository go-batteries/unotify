run:
	../build/terrawrapper.sh validate
	../build/terrawrapper.sh plan -out ./tfstates/dashsotdash.tfplan
	../build/terrawrapper.sh apply "./tfstates/dashsotdash.tfplan"
	chmod 600 dash-tf-key-pair.pem

destroy:
	../build/terrawrapper.sh destroy "./tfstates/dashsotdash.tfplan"

