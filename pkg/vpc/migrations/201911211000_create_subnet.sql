CREATE TABLE subnets (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	vpc_id INT NOT NULL,
	region STRING NOT NULL,
	cidr INET NOT NULL,
	inner_vlan INT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	metadata JSONB
);