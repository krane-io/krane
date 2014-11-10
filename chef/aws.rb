require 'chef_metal'
require 'chef_metal_fog'
with_chef_local_server :chef_repo_path => "#{File.dirname(__FILE__)}", :cookbook_path => "#{File.dirname(__FILE__)}/cookbooks"

to_exit = false
configuration = Hash.new
configuration['aws_machine_name'] 		= ENV['AWS_MACHINE_NAME']
configuration['aws_plan_id'] 			= ENV['AWS_PLAN_ID']
configuration['aws_access_key_id']		= ENV['AWS_ACCESS_KEY_ID']
configuration['aws_secret_access_key']	= ENV['AWS_SECRET_ACCESS_KEY']
configuration['aws_key']                = ENV['AWS_KEY']


configuration.each do |key, value|
	if value.nil?
		puts " + Please insert value in your enviroment key #{key.upcase}"
		to_exit = true
	end
end

exit 1 if to_exit


plan = {
	"53f0f0ecd8a5975a1c000162" => {
		:region => "us-east-1a",
		:flavour => "m1.small"
	},
	"53f0f0ecd8a5975a1c000163" => {
		:region => "us-east-1a",
		:flavour => "m1.medium"
	},
	"53f0f0ecd8a5975a1c000164" => {
		:region => "us-east-1a",
		:flavour => "m1.large"
	},
	"53f0f0ecd8a5975a1c000165" => {
		:region => "us-east-1a",
		:flavour => "m1.xlarge"
	},
	"53f0f0ecd8a5975a1c000166" => {
		:region => "us-east-1a",
		:flavour => "t1.micro"
	},
	"53f0f0ecd8a5975a1c000167" => {
		:region => "us-east-1a",
		:flavour => "m2.xlarge"
	},
	"53f0f0ecd8a5975a1c000168" => {
		:region => "us-east-1a",
		:flavour => "m2.2xlarge"
	},
	"53f0f0ecd8a5975a1c000169" => {
		:region => "us-east-1a",
		:flavour => "m2.4xlarge"
	},
	"53f0f0ecd8a5975a1c00016a" => {
		:region => "us-east-1a",
		:flavour => "c1.medium"
	},
	"53f0f0ecd8a5975a1c00016b" => {
		:region => "us-east-1a",
		:flavour => "c1.xlarge"
	},
	"53f0f0ecd8a5975a1c00016c" => {
		:region => "us-east-1a",
		:flavour => "m3.xlarge"
	},
	"53f0f0ecd8a5975a1c00016d" => {
		:region => "us-east-1a",
		:flavour => "m3.2xlarge"
	},
	"53f0f0ecd8a5975a1c00016e" => {
		:region => "us-east-1b",
		:flavour => "m1.small"
	},
	"53f0f0ecd8a5975a1c00016f" => {
		:region => "us-east-1b",
		:flavour => "m1.medium"
	},
	"53f0f0ecd8a5975a1c000170" => {
		:region => "us-east-1b",
		:flavour => "m1.large"
	},
	"53f0f0ecd8a5975a1c000171" => {
		:region => "us-east-1b",
		:flavour => "m1.xlarge"
	},
	"53f0f0ecd8a5975a1c000172" => {
		:region => "us-east-1b",
		:flavour => "t1.micro"
	},
	"53f0f0ecd8a5975a1c000173" => {
		:region => "us-east-1b",
		:flavour => "m2.xlarge"
	},
	"53f0f0ecd8a5975a1c000174" => {
		:region => "us-east-1b",
		:flavour => "m2.2xlarge"
	},
	"53f0f0ecd8a5975a1c000175" => {
		:region => "us-east-1b",
		:flavour => "m2.4xlarge"
	},
	"53f0f0ecd8a5975a1c000176" => {
		:region => "us-east-1b",
		:flavour => "c1.medium"
	},
	"53f0f0ecd8a5975a1c000177" => {
		:region => "us-east-1b",
		:flavour => "c1.xlarge"
	},
	"53f0f0ecd8a5975a1c000178" => {
		:region => "us-east-1b",
		:flavour => "m3.xlarge"
	},
	"53f0f0ecd8a5975a1c000179" => {
		:region => "us-east-1b",
		:flavour => "m3.2xlarge"
	},
	"53f0f0ecd8a5975a1c00017a" => {
		:region => "us-east-1c",
		:flavour => "m1.small"
	},
	"53f0f0ecd8a5975a1c00017b" => {
		:region => "us-east-1c",
		:flavour => "m1.medium"
	},
	"53f0f0ecd8a5975a1c00017c" => {
		:region => "us-east-1c",
		:flavour => "m1.large"
	},
	"53f0f0ecd8a5975a1c00017d" => {
		:region => "us-east-1c",
		:flavour => "m1.xlarge"
	},
	"53f0f0ecd8a5975a1c00017e" => {
		:region => "us-east-1c",
		:flavour => "t1.micro"
	},
	"53f0f0ecd8a5975a1c00017f" => {
		:region => "us-east-1c",
		:flavour => "m2.xlarge"
	},
	"53f0f0ecd8a5975a1c000180" => {
		:region => "us-east-1c",
		:flavour => "m2.2xlarge"
	},
	"53f0f0ecd8a5975a1c000181" => {
		:region => "us-east-1c",
		:flavour => "m2.4xlarge"
	},
	"53f0f0ecd8a5975a1c000182" => {
		:region => "us-east-1c",
		:flavour => "c1.medium"
	},
	"53f0f0ecd8a5975a1c000183" => {
		:region => "us-east-1c",
		:flavour => "c1.xlarge"
	},
	"53f0f0ecd8a5975a1c000184" => {
		:region => "us-east-1c",
		:flavour => "m3.xlarge"
	},
	"53f0f0ecd8a5975a1c000185" => {
		:region => "us-east-1c",
		:flavour => "m3.2xlarge"
	},
	"53f0f0ecd8a5975a1c000186" => {
		:region => "us-east-1d",
		:flavour => "m1.small"
	},
	"53f0f0ecd8a5975a1c000187" => {
		:region => "us-east-1d",
		:flavour => "m1.medium"
	},
	"53f0f0ecd8a5975a1c000188" => {
		:region => "us-east-1d",
		:flavour => "m1.large"
	},
	"53f0f0ecd8a5975a1c000189" => {
		:region => "us-east-1d",
		:flavour => "m1.xlarge"
	},
	"53f0f0ecd8a5975a1c00018a" => {
		:region => "us-east-1d",
		:flavour => "t1.micro"
	},
	"53f0f0ecd8a5975a1c00018b" => {
		:region => "us-east-1d",
		:flavour => "m2.xlarge"
	},
	"53f0f0ecd8a5975a1c00018c" => {
		:region => "us-east-1d",
		:flavour => "m2.2xlarge"
	},
	"53f0f0ecd8a5975a1c00018d" => {
		:region => "us-east-1d",
		:flavour => "m2.4xlarge"
	},
	"53f0f0ecd8a5975a1c00018e" => {
		:region => "us-east-1d",
		:flavour => "c1.medium"
	},
	"53f0f0ecd8a5975a1c00018f" => {
		:region => "us-east-1d",
		:flavour => "c1.xlarge"
	},
	"53f0f0ecd8a5975a1c000190" => {
		:region => "us-east-1d",
		:flavour => "m3.xlarge"
	},
	"53f0f0ecd8a5975a1c000191" => {
		:region => "us-east-1d",
		:flavour => "m3.2xlarge"
	},
	"53f0f0ecd8a5975a1c000192" => {
		:region => "us-east-1e",
		:flavour => "m1.small"
	},
	"53f0f0ecd8a5975a1c000193" => {
		:region => "us-east-1e",
		:flavour => "m1.medium"
	},
	"53f0f0ecd8a5975a1c000194" => {
		:region => "us-east-1e",
		:flavour => "m1.large"
	},
	"53f0f0ecd8a5975a1c000195" => {
		:region => "us-east-1e",
		:flavour => "m1.xlarge"
	},
	"53f0f0ecd8a5975a1c000196" => {
		:region => "us-east-1e",
		:flavour => "t1.micro"
	},
	"53f0f0ecd8a5975a1c000197" => {
		:region => "us-east-1e",
		:flavour => "m2.xlarge"
	},
	"53f0f0ecd8a5975a1c000198" => {
		:region => "us-east-1e",
		:flavour => "m2.2xlarge"
	},
	"53f0f0ecd8a5975a1c000199" => {
		:region => "us-east-1e",
		:flavour => "m2.4xlarge"
	},
	"53f0f0ecd8a5975a1c00019a" => {
		:region => "us-east-1e",
		:flavour => "c1.medium"
	},
	"53f0f0ecd8a5975a1c00019b" => {
		:region => "us-east-1e",
		:flavour => "c1.xlarge"
	},
	"53f0f0ecd8a5975a1c00019c" => {
		:region => "us-east-1e",
		:flavour => "m3.xlarge"
	},
	"53f0f0ecd8a5975a1c00019d" => {
		:region => "us-east-1e",
		:flavour => "m3.2xlarge"
	},
	"53f0f0ecd8a5975a1c00019e" => {
		:region => "eu-west-1a",
		:flavour => "m1.small"
	},
	"53f0f0ecd8a5975a1c00019f" => {
		:region => "eu-west-1a",
		:flavour => "m1.medium"
	},
	"53f0f0ecd8a5975a1c0001a0" => {
		:region => "eu-west-1a",
		:flavour => "m1.large"
	},
	"53f0f0ecd8a5975a1c0001a1" => {
		:region => "eu-west-1a",
		:flavour => "m1.xlarge"
	},
	"53f0f0ecd8a5975a1c0001a2" => {
		:region => "eu-west-1a",
		:flavour => "t1.micro"
	},
	"53f0f0ecd8a5975a1c0001a3" => {
		:region => "eu-west-1a",
		:flavour => "m2.xlarge"
	},
	"53f0f0ecd8a5975a1c0001a4" => {
		:region => "eu-west-1a",
		:flavour => "m2.2xlarge"
	},
	"53f0f0ecd8a5975a1c0001a5" => {
		:region => "eu-west-1a",
		:flavour => "m2.4xlarge"
	},
	"53f0f0ecd8a5975a1c0001a6" => {
		:region => "eu-west-1a",
		:flavour => "c1.medium"
	},
	"53f0f0ecd8a5975a1c0001a7" => {
		:region => "eu-west-1a",
		:flavour => "c1.xlarge"
	},
	"53f0f0ecd8a5975a1c0001a8" => {
		:region => "eu-west-1b",
		:flavour => "m1.small"
	},
	"53f0f0ecd8a5975a1c0001a9" => {
		:region => "eu-west-1b",
		:flavour => "m1.medium"
	},
	"53f0f0ecd8a5975a1c0001aa" => {
		:region => "eu-west-1b",
		:flavour => "m1.large"
	},
	"53f0f0ecd8a5975a1c0001ab" => {
		:region => "eu-west-1b",
		:flavour => "m1.xlarge"
	},
	"53f0f0ecd8a5975a1c0001ac" => {
		:region => "eu-west-1b",
		:flavour => "t1.micro"
	},
	"53f0f0ecd8a5975a1c0001ad" => {
		:region => "eu-west-1b",
		:flavour => "m2.xlarge"
	},
	"53f0f0ecd8a5975a1c0001ae" => {
		:region => "eu-west-1b",
		:flavour => "m2.2xlarge"
	},
	"53f0f0ecd8a5975a1c0001af" => {
		:region => "eu-west-1b",
		:flavour => "m2.4xlarge"
	},
	"53f0f0ecd8a5975a1c0001b0" => {
		:region => "eu-west-1b",
		:flavour => "c1.medium"
	},
	"53f0f0ecd8a5975a1c0001b1" => {
		:region => "eu-west-1b",
		:flavour => "c1.xlarge"
	},
	"53f0f0ecd8a5975a1c0001b2" => {
		:region => "eu-west-1c",
		:flavour => "m1.small"
	},
	"53f0f0ecd8a5975a1c0001b3" => {
		:region => "eu-west-1c",
		:flavour => "m1.medium"
	},
	"53f0f0ecd8a5975a1c0001b4" => {
		:region => "eu-west-1c",
		:flavour => "m1.large"
	},
	"53f0f0ecd8a5975a1c0001b5" => {
		:region => "eu-west-1c",
		:flavour => "m1.xlarge"
	},
	"53f0f0ecd8a5975a1c0001b6" => {
		:region => "eu-west-1c",
		:flavour => "t1.micro"
	},
	"53f0f0ecd8a5975a1c0001b7" => {
		:region => "eu-west-1c",
		:flavour => "m2.xlarge"
	},
	"53f0f0ecd8a5975a1c0001b8" => {
		:region => "eu-west-1c",
		:flavour => "m2.2xlarge"
	},
	"53f0f0ecd8a5975a1c0001b9" => {
		:region => "eu-west-1c",
		:flavour => "m2.4xlarge"
	},
	"53f0f0ecd8a5975a1c0001ba" => {
		:region => "eu-west-1c",
		:flavour => "c1.medium"
	},
	"53f0f0ecd8a5975a1c0001bb" => {
		:region => "eu-west-1c",
		:flavour => "c1.xlarge"
	},
	"53f0f0ecd8a5975a1c0001bc" => {
		:region => "us-west-1a",
		:flavour => "m1.small"
	},
	"53f0f0ecd8a5975a1c0001bd" => {
		:region => "us-west-1a",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c0001be" => {
		:region => "us-west-1a",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c0001bf" => {
		:region => "us-west-1a",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c0001c0" => {
		:region => "us-west-1a",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c0001c1" => {
		:region => "us-west-1a",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c0001c2" => {
		:region => "us-west-1a",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c0001c3" => {
		:region => "us-west-1a",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c0001c4" => {
		:region => "us-west-1a",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c0001c5" => {
		:region => "us-west-1a",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c0001c6" => {
		:region => "us-west-1b",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c0001c7" => {
		:region => "us-west-1b",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c0001c8" => {
		:region => "us-west-1b",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c0001c9" => {
		:region => "us-west-1b",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c0001ca" => {
		:region => "us-west-1b",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c0001cb" => {
		:region => "us-west-1b",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c0001cc" => {
		:region => "us-west-1b",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c0001cd" => {
		:region => "us-west-1b",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c0001ce" => {
		:region => "us-west-1b",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c0001cf" => {
		:region => "us-west-1b",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c0001d0" => {
		:region => "us-west-1c",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c0001d1" => {
		:region => "us-west-1c",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c0001d2" => {
		:region => "us-west-1c",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c0001d3" => {
		:region => "us-west-1c",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c0001d4" => {
		:region => "us-west-1c",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c0001d5" => {
		:region => "us-west-1c",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c0001d6" => {
		:region => "us-west-1c",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c0001d7" => {
		:region => "us-west-1c",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c0001d8" => {
		:region => "us-west-1c",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c0001d9" => {
		:region => "us-west-1c",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c0001da" => {
		:region => "ap-southeast-1a",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c0001db" => {
		:region => "ap-southeast-1a",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c0001dc" => {
		:region => "ap-southeast-1a",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c0001dd" => {
		:region => "ap-southeast-1a",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c0001de" => {
		:region => "ap-southeast-1a",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c0001df" => {
		:region => "ap-southeast-1a",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c0001e0" => {
		:region => "ap-southeast-1a",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c0001e1" => {
		:region => "ap-southeast-1a",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c0001e2" => {
		:region => "ap-southeast-1a",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c0001e3" => {
		:region => "ap-southeast-1a",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c0001e4" => {
		:region => "ap-southeast-1b",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c0001e5" => {
		:region => "ap-southeast-1b",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c0001e6" => {
		:region => "ap-southeast-1b",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c0001e7" => {
		:region => "ap-southeast-1b",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c0001e8" => {
		:region => "ap-southeast-1b",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c0001e9" => {
		:region => "ap-southeast-1b",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c0001ea" => {
		:region => "ap-southeast-1b",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c0001eb" => {
		:region => "ap-southeast-1b",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c0001ec" => {
		:region => "ap-southeast-1b",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c0001ed" => {
		:region => "ap-southeast-1b",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c0001ee" => {
		:region => "ap-northeast-1a",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c0001ef" => {
		:region => "ap-northeast-1a",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c0001f0" => {
		:region => "ap-northeast-1a",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c0001f1" => {
		:region => "ap-northeast-1a",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c0001f2" => {
		:region => "ap-northeast-1a",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c0001f3" => {
		:region => "ap-northeast-1a",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c0001f4" => {
		:region => "ap-northeast-1a",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c0001f5" => {
		:region => "ap-northeast-1a",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c0001f6" => {
		:region => "ap-northeast-1a",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c0001f7" => {
		:region => "ap-northeast-1a",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c0001f8" => {
		:region => "ap-northeast-1b",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c0001f9" => {
		:region => "ap-northeast-1b",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c0001fa" => {
		:region => "ap-northeast-1b",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c0001fb" => {
		:region => "ap-northeast-1b",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c0001fc" => {
		:region => "ap-northeast-1b",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c0001fd" => {
		:region => "ap-northeast-1b",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c0001fe" => {
		:region => "ap-northeast-1b",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c0001ff" => {
		:region => "ap-northeast-1b",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c000200" => {
		:region => "ap-northeast-1b",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c000201" => {
		:region => "ap-northeast-1b",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c000202" => {
		:region => "ap-northeast-1c",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c000203" => {
		:region => "ap-northeast-1c",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c000204" => {
		:region => "ap-northeast-1c",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c000205" => {
		:region => "ap-northeast-1c",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c000206" => {
		:region => "ap-northeast-1c",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c000207" => {
		:region => "ap-northeast-1c",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c000208" => {
		:region => "ap-northeast-1c",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c000209" => {
		:region => "ap-northeast-1c",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c00020a" => {
		:region => "ap-northeast-1c",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c00020b" => {
		:region => "ap-northeast-1c",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c00020c" => {
		:region => "us-west-2a",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c00020d" => {
		:region => "us-west-2a",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c00020e" => {
		:region => "us-west-2a",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c00020f" => {
		:region => "us-west-2a",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c000210" => {
		:region => "us-west-2a",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c000211" => {
		:region => "us-west-2a",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c000212" => {
		:region => "us-west-2a",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c000213" => {
		:region => "us-west-2a",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c000214" => {
		:region => "us-west-2a",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c000215" => {
		:region => "us-west-2a",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c000216" => {
		:region => "us-west-2b",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c000217" => {
		:region => "us-west-2b",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c000218" => {
		:region => "us-west-2b",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c000219" => {
		:region => "us-west-2b",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c00021a" => {
		:region => "us-west-2b",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c00021b" => {
		:region => "us-west-2b",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c00021c" => {
		:region => "us-west-2b",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c00021d" => {
		:region => "us-west-2b",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c00021e" => {
		:region => "us-west-2b",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c00021f" => {
		:region => "us-west-2b",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c000220" => {
		:region => "us-west-2c",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c000221" => {
		:region => "us-west-2c",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c000222" => {
		:region => "us-west-2c",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c000223" => {
		:region => "us-west-2c",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c000224" => {
		:region => "us-west-2c",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c000225" => {
		:region => "us-west-2c",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c000226" => {
		:region => "us-west-2c",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c000227" => {
		:region => "us-west-2c",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c000228" => {
		:region => "us-west-2c",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c000229" => {
		:region => "us-west-2c",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c00022a" => {
		:region => "sa-east-1a",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c00022b" => {
		:region => "sa-east-1a",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c00022c" => {
		:region => "sa-east-1a",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c00022d" => {
		:region => "sa-east-1a",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c00022e" => {
		:region => "sa-east-1a",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c00022f" => {
		:region => "sa-east-1a",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c000230" => {
		:region => "sa-east-1a",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c000231" => {
		:region => "sa-east-1a",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c000232" => {
		:region => "sa-east-1a",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c000233" => {
		:region => "sa-east-1a",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c000234" => {
		:region => "sa-east-1b",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c000235" => {
		:region => "sa-east-1b",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c000236" => {
		:region => "sa-east-1b",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c000237" => {
		:region => "sa-east-1b",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c000238" => {
		:region => "sa-east-1b",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c000239" => {
		:region => "sa-east-1b",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c00023a" => {
		:region => "sa-east-1b",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c00023b" => {
		:region => "sa-east-1b",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c00023c" => {
		:region => "sa-east-1b",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c00023d" => {
		:region => "sa-east-1b",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c00023e" => {
		:region => "ap-southeast-2a",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c00023f" => {
		:region => "ap-southeast-2a",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c000240" => {
		:region => "ap-southeast-2a",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c000241" => {
		:region => "ap-southeast-2a",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c000242" => {
		:region => "ap-southeast-2a",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c000243" => {
		:region => "ap-southeast-2a",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c000244" => {
		:region => "ap-southeast-2a",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c000245" => {
		:region => "ap-southeast-2a",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c000246" => {
		:region => "ap-southeast-2a",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c000247" => {
		:region => "ap-southeast-2a",
		:flavour => "c1.xlarge"
	},
	"53f0f0edd8a5975a1c000248" => {
		:region => "ap-southeast-2b",
		:flavour => "m1.small"
	},
	"53f0f0edd8a5975a1c000249" => {
		:region => "ap-southeast-2b",
		:flavour => "m1.medium"
	},
	"53f0f0edd8a5975a1c00024a" => {
		:region => "ap-southeast-2b",
		:flavour => "m1.large"
	},
	"53f0f0edd8a5975a1c00024b" => {
		:region => "ap-southeast-2b",
		:flavour => "m1.xlarge"
	},
	"53f0f0edd8a5975a1c00024c" => {
		:region => "ap-southeast-2b",
		:flavour => "t1.micro"
	},
	"53f0f0edd8a5975a1c00024d" => {
		:region => "ap-southeast-2b",
		:flavour => "m2.xlarge"
	},
	"53f0f0edd8a5975a1c00024e" => {
		:region => "ap-southeast-2b",
		:flavour => "m2.2xlarge"
	},
	"53f0f0edd8a5975a1c00024f" => {
		:region => "ap-southeast-2b",
		:flavour => "m2.4xlarge"
	},
	"53f0f0edd8a5975a1c000250" => {
		:region => "ap-southeast-2b",
		:flavour => "c1.medium"
	},
	"53f0f0edd8a5975a1c000251" => {
		:region => "ap-southeast-2b",
		:flavour => "c1.xlarge"
	}
}


with_driver 'fog:AWS', :compute_options => {
	:aws_access_key_id => configuration['aws_access_key_id'],
	:aws_secret_access_key => configuration['aws_secret_access_key'],
	:region => plan[configuration['aws_plan_id']][:region].chop
}




fog_key_pair 'id_rsa' do
	private_key_path "#{File.expand_path('~')}/.ssh/#{configuration['aws_key']}"
	public_key_path "#{File.expand_path('~')}/.ssh/#{configuration['aws_key']}.pub"
	allow_overwrite true
end

#
# Ubuntu 14.04 Oficial AMIs for AWS. List obtained from: http://cloud-images.ubuntu.com/locator/ec2/
#
amis = Hash.new
amis['ap-northeast-1'] 	= 'ami-d54b60d4'
amis['ap-southeast-1'] 	= 'ami-24e7c076'
amis['eu-west-1'] 		= 'ami-00b11177'
amis['sa-east-1'] 		= 'ami-79d26764'
amis['us-east-1'] 		= 'ami-8caa1ce4'
amis['us-west-1'] 		= 'ami-696e652c'
amis['cn-north-1'] 		= 'ami-9e42d0a7'
amis['us-gov-west-1'] 	= 'ami-f34d2ad0'
amis['ap-southeast-2'] 	= 'ami-2111731b'
amis['us-west-2'] 		= 'ami-cd5311fd'


with_machine_options :bootstrap_options => {
	:key_name => 'id_rsa',
	:image_id => amis[plan[configuration['aws_plan_id']][:region].chop],
	:flavor_id => plan[configuration['aws_plan_id']][:flavour],
	:tags => {"docker" => 'true'}
}


machine configuration['aws_machine_name'] do
	tag 'docker'
	recipe 'docker'
	recipe 'cadvisor'
	recipe 'openssh'
	attributes(
		"docker" => {
			"docker_daemon_timeout" => 30,
			"host" => ['tcp://localhost:27017', 'unix:///var/run/docker.sock']
		},
		"cadvisor" => {
			"listen_ip" => "0.0.0.0"
		},
		"openssh" => {
			"server" => {
					"permit_root_login" => "yes"
			},
			"disable_root_opts" => true
		}
	)

	converge true
end
