run_plan_debug:
	go install . && TF_LOG=DEBUG terraform plan

run_plan_info:
	go install . && TF_LOG=INFO terraform plan

run_apply_debug:
	go install . && TF_LOG=DEBUG terraform apply

run_apply_info:
	go install . && TF_LOG=INFO terraform apply