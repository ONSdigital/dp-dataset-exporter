package main

import (
	"bufio"
	"context"
	"os"
	"time"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/log"
)

// The purpose of this app is to trigger various filters in order to time how long they take
// my examining the logs

func main() {
	ctx := context.Background()
	log.Namespace = "dp-dataset-exporter"

	config, err := config.Get()
	if err != nil {
		log.Event(ctx, "error getting config", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	// Avoid logging the neo4j FileURL as it may contain a password
	log.Event(ctx, "loaded config", log.INFO, log.Data{"config": config})

	// Create Kafka Producer
	pChannels := kafka.CreateProducerChannels()
	pConfig := &kafka.ProducerConfig{KafkaVersion: &config.KafkaVersion}
	kafkaProducer, err := kafka.NewProducer(ctx, config.KafkaAddr, config.FilterConsumerTopic, pChannels, pConfig)
	if err != nil {
		log.Event(ctx, "fatal error trying to create kafka producer", log.FATAL, log.Error(err), log.Data{"topic": config.FilterConsumerTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.Channels().LogErrors(ctx, "kafka producer")

	filterIds := []string{
		"08eda1d1-df67-499a-a545-b5223292b19f", // this works ?
		//		"ecb98b95-d24d-4083-8c71-ee95d51b8806",
		//		"fc6e0972-e332-44e1-abd3-02b9a52e4d8a",
		//		"b975e8d7-04e6-4405-bd3e-ad5c5dab9984",
		//		"937f6db3-ad42-48c3-b37f-ebb787ab1b37",
		//		"7e060612-11f5-474c-b7d2-90659dbda44f",
		"460b5039-bb09-4038-b8eb-9091713f4497", // this works ?
		"8b6e93bb-d40d-4370-b75f-a25aecf9fb7f", // this works ?
		//		"b81aae94-1ace-4b8b-967d-e9e0db98a1bd",
		//		"badd7b3a-ab9c-4f4c-b226-9b4c87f7f68e",
		//		"6ebfabc2-03d4-4b5f-8347-1b9a69307380",
		//		"a1845d73-8c42-4d6e-816e-cd3b990047c7",
		//		"38ba0eab-5d28-4f35-9c57-bcb78eb9525d",
		//		"f4c710c7-f8ab-4873-9aff-aa2ab205fd08",
		//		"6f2bc166-5a4d-4121-880b-13afcf6bd74b",
		//		"e72945c2-61a2-4769-8ced-e47826baf482",
		//		"30c368ec-ea03-4845-b884-6260180ea5d2",
		//		"fa8f16ef-ded5-468a-85f3-9145eef4f184",
		//		"55ee5f56-d8a5-48ea-be3e-fb4bed61beaf",
		//		"590ac1d6-4891-4ac6-97d7-37ea0dd546e4",
		//		"dd9400d4-d11f-4b13-95ed-8a0e596da067",
		//		"2af49bf9-5a18-4ab9-a931-0bea4688dba3",
		//		"d9bde796-e90a-4f28-a8f3-dfb74c087184",
		//		"7a8a953b-04cc-4ac7-a27e-b637fa8c8e2b",
		//		"572e9a31-3438-41fa-8e45-06f6551f0a3e",
		//		"38160308-39a4-4cd1-b999-7e9c61834c44",
		//		"e134c9ec-74f1-4040-a7c3-39ebcc1350ee",
		//		"39c67fc5-0ddf-4361-89e3-77a5a337af2b",
		//		"72aa7da7-c8ba-4bf5-9deb-19ba79c66b4e",
		//		"dfbf5fbf-6be2-4531-a590-a8737e5f0e51",
		//		"011b14a9-bcad-4ac2-a931-2a4d46377b65",
		//		"b41b847e-a873-4e54-8544-582c56852040",
		//		"e95baa98-e524-41f6-a5b9-d06f39b613c7",
		//		"5b8a04e4-4348-44cf-a42b-85aeae380afa",
		//		"65efe534-6955-4f67-86b4-0023a78575c3",
		//		"45db88f8-5fb0-4770-871b-feef199e015a",
		//		"f8cadc38-03df-4e0b-a9cf-a3806adeef4e",
		//		"a9b8ad38-c8a8-4cb3-ad66-f3bd2999bc26",
		//		"782c8053-e095-4454-a48b-b61f7736f4fa",
		//		"74cce473-71f0-4a72-b6b1-c406ed02a2a9",
		//		"7ef109a5-97f0-4c3d-849e-de8f1c2efa43",
		//		"a78e5f8c-adbb-47d3-a1ce-10e49c00c8f0",
		//		"a420f92d-60fa-4ffd-8ee6-90a914a07b8c",
		//		"02db7be1-d2d9-45e7-9f69-25a28119a5bd",
		//		"4ac9a0a9-8d03-4c9d-b732-efef04d99395",
		//		"095f3d15-97c0-432c-b7c7-e459d9a0fd38",
		//		"1e16b768-af48-4969-a9b5-cd7db114e0f2",
		//		"12aac449-52d7-4cbd-97f4-590937cf90fd",
		//		"51e48fb8-16bc-49b8-a468-5d500e3c59a9",
		//		"a69acc7b-b654-4811-9704-27d7e88255aa",
		//		"94de0fa6-525e-4744-a447-7127dd5e3221",
		//		"a39d0b84-9937-4ea7-b659-6c8e8e1234f8",
		//		"4d9b66d0-1041-4b58-a7a9-5d7c246b189f",
		//		"44123343-3a11-4ff6-bc41-8364f71dea6f",
		//		"382a9902-fd36-465e-a0e4-c0553ce14a4b",
		//		"bc2c3243-490d-4a35-bc4f-89d02c4667a9",
		//		"d634a4f2-728c-4a84-89bf-1454404313f3",
		//		"c83ab8ff-bb82-4cba-b3a8-e1630fbe839a",
		//		"20ceda0e-dfc2-4ff3-b110-2b528418c1c0",
		//		"8cb42181-e596-4617-b158-08fb515bcc09",
		//		"bee0ad53-b366-484e-8418-8d2666dce03b",
		//		"1d61ba76-f0d9-41b0-995e-681bee719294",
		//		"d983f8ff-7c65-4d71-8e3a-2bd83df38353",
		//		"dc1ccecb-c8dd-4f6a-9f7c-5786a1f0cc44", // this fail
		//		"53affc45-1d12-4652-acdd-9970cc7dfc87", // this fail
		//		"6a21b2e8-4b90-45ab-a7d8-dace66328ad5", // this fail
		//		"33908f59-56fa-4b6e-bea5-9a49783de7ca", // this fail
		//		"2f7f8e17-168a-4645-8b11-8f9c5773eca6", // this fail
		//		"0412d7de-5a1f-4395-955f-b2be52e42f12", // this fail
		//		"30473f62-5ace-44e3-8553-d2f40441f004", // this fail
		//		"6cfc92d4-8d47-41e0-99e8-79df48fce83b",
		//		"04e6fee3-ba14-4565-b8ec-51b58dc4b193",
		//		"2f6adc38-faaf-4c6c-97ee-01a78604869f",
		//		"dc5bceba-aebc-4fc3-9501-966b1b8849e2",
		//		"88600f4b-aff2-42ce-baeb-187f1a5c4297",
		//		"17ef57c4-953a-4f00-8a50-d49fd08487b0",
		//		"b6f1262a-735f-4cd1-9ea9-da9774d5e4d2",
		//		"ccc5b75c-9cc6-4293-afcc-7ab30c2baa53",
		//		"2e1fcaf9-2518-4fb0-9b4c-eca4b16e4dc4",
		//		"7b942d4b-a047-4665-8e08-c3d73e78473d",
		//		"6c549fdf-f643-4909-924f-81ad8fdcbd82",
		//		"e742246a-9acb-4743-b205-abbd50645455",
		//		"f8ae0852-88e9-46a0-a0c8-616c6edf4a8d",
		//		"86bfdc4c-87ff-4e2a-adb2-ce398d387790",
		//		"cc8c6c3c-52bc-4657-8581-7661fc89ba8e",
		//		"cecd2910-1098-4e1b-b85d-3f50d1e23172",
		//		"2af26a2c-7434-4a3f-bb01-c1ec84439238",
		//		"2dd01b68-f2af-40ea-97e8-b468d9e52959",
		//		"bbe149b2-66a4-4c3a-bc91-cfad90b0d3e5",
		//		"292ede17-a216-402d-9681-b20781210a2c",
		//		"604d1046-d313-450b-b796-78376ea874ec",
		//		"1c007f2c-a70e-4645-a8c1-c1e1a005c91e",
		//		"9b0abb36-f680-4ee0-b6a2-b29b31ef09c9",
		//		"3d6efa71-5e7c-4128-a0ba-fcb445294976",
		//		"1f3eb40a-3e45-43e1-8e02-6442190ac6f5",
		//		"df586e6a-1f25-466b-8fe8-056b3be58740",
		//		"5884367b-c415-49b7-a21f-9c44e8d1c16e",
		//		"9332b945-fac2-4e77-b1d1-8741ef3e695e",
		//		"287c6d86-157e-4253-a848-ef15ce409418",
		//		"16eeaea7-bc59-4098-b301-ac92564c88f5",
		//		"a7b465a5-311f-444e-926e-d2d9769b9c56",
		//		"0cf56e44-575c-4978-987c-00176f704ff8",
		//		"edb47bdb-8f6c-4317-8c8d-1c2fd9cfdeaf",
		//		"5541d5f9-8eaf-4ff5-8e0b-9f62d9e35e8a",
		//		"ca606df7-d432-4fa1-ad9a-701901fdfa03",
		//		"9d395d45-01f8-4051-8e8a-01345c40f2fa",
		//		"ef7c8693-df0c-4dfe-8f79-1978301bfb09",
		//		"dd47bd7c-43d1-4c17-ba6b-9dc279ba128a",
		//		"105a43e6-f327-4aab-8afd-957da1f91823",
		//		"9380bc95-7a68-4f44-95a6-149df2466c95",
		//		"f5dc140a-c842-4126-ba9a-e49d0b1d74aa",
		//		"36d4c129-de6c-4e91-b9aa-15b1e8a4b797",
		//		"08da9747-fbef-4d53-89c7-4ccd22613f5b",
		//		"f2d8ce18-6704-4860-8d9f-b6dc6c7066d4",
		//		"fdb58696-de04-499b-abeb-6258607d4ba6",
		//		"16b7daed-4a2c-4de8-8df9-ae17a29dd63b",
		//		"068935eb-6ea1-4c03-88a3-bb9e25fd888f",
		//		"de0aff62-9b5c-47ba-abfd-db666afc0341",
		//		"35539f42-dee8-4b1d-91dd-a1e1fcc2849e",
		//		"c5f5192f-3424-4de8-9fb5-aff16fce2dea",
		//		"e530822b-8c54-4fec-a7d8-ae3ad7904ff8",
		//		"0de92eb4-206f-4636-8ce4-ea9d6463183b",
		//		"9b857212-efbc-4f24-967b-ee3916bfe95d",
		//		"f6c90679-7b82-4950-a861-602248c84378",
		//		"cc893af7-3a40-420b-b657-34d808cd6a00",
		//		"9d91e9fb-737d-4c42-af79-115e9a51d1e5",
		//		"4368a4fa-18b6-4736-9da0-15379abaf0ea",
		//		"02d8cfcb-f4a7-4ba2-8520-8286d1e886fe",
		//		"8cf71dab-9a76-43c2-a615-a5425688e079",
		//		"9a439f4a-e677-4276-8b31-88be36f2fd90",
		//		"a1eb7f8f-9519-4ecd-bcf9-aa355646a6a4",
		//		"26ed0c7a-0186-4071-b5c5-0277f9e0c57b",
		//		"a933839d-fc85-4a0b-82f4-662553ce4d6f",
		//		"1b2a74cc-fa8f-468d-bf56-847dffce81c8",
		//		"a91cfe98-3d23-4dd9-833a-f925b2326317",
		"c3a5bf00-c975-46f9-83e9-748f5076e47a", // this works ? ~ 30 seconds
		//		"45c03551-cc05-48d4-a725-c7a4e2fd5ff9",
		//		"4f19565b-a601-481e-8ad7-39b400d70c32",
		//		"7e11cd60-3be1-4e5d-9049-79361eba390b",
		//		"409dd7e9-8e4d-4a00-89ba-1f66f0e66e3e",
		//		"340dae08-fd75-48a5-80b4-4e09cf5cf80f",
		//		"51799325-1870-45cf-b45e-d756448dc3bb",
		//		"b1826e42-75ab-4eba-8339-86a1df00a63d",
		//		"fa7a46e7-67d9-4c62-b21d-796a4d34d5cc",
		//		"d2ea1bb5-9363-4813-88d4-abc90ff85c81",
		//		"c8cdf441-53e6-4e69-a061-18b2b09ed39f",
		//		"6fc14fdf-60c4-4dc6-ba7a-4292a888c07f",
		//		"1bd4b374-73ea-408e-ae46-546913caa534",
		//		"0880d82b-39e1-4792-a11e-736c4dba5260",
		//		"117f12d8-04f8-4532-977d-f52eb2393513",
		"d49aaa19-1daa-43d4-9861-dbe88d5b9d54", // this works ?
		//		"bc2c2d6d-c62e-42fc-a8b7-a2ab4688dd53",
		//		"0ae5f4d4-34d7-4e69-b8d1-088f63ae22aa",
		//		"ba42ad8c-c8cc-435e-b9bf-bbe41691311b",
		//		"2ae58afe-2ea6-40f5-8211-f225a4858d58",
		//		"92790efd-2807-42c6-98fe-04ba7aae06a4",
		//		"6afe9c81-752b-4e4b-a121-ba2350fe174f",
		//		"e861634f-5041-4bc5-9d5b-58585aba37d9",
		//		"be2798ca-f4d6-491b-b8d2-4cd91c1520ef",
		//		"59733ce9-1f48-4056-8aae-d11f5cc3bbfe",
		//		"bc422beb-6bc2-433e-b0ac-656ddbe03fcd",
		//		"6346673f-4631-4ad4-9403-a72731fc66e6",
		//		"178b0d2a-f536-4d8b-ad9d-e6812ee16f8a",
		//		"e736f37c-57fc-47c9-8707-cd455d051b5a",
		//		"5e375606-2568-45ec-9573-364f45e4619d",
		//		"1b3cdd4f-d4fd-4def-a9b7-3d23e0fb5970",
		//		"394f369c-674d-4b71-b84f-ba24e3371d0c",
		//		"8a62deee-19c5-4c07-bce7-6431917d5482",
		//		"2292400f-509e-4983-bd7b-7dca6ff67d8f",
		//		"b095e225-3641-4a46-b6ce-f273d4a6a70b",
		//		"ea79c9d3-6fab-4b6d-97a2-00a5c8b9d0ac",
		//		"b134a412-1dbe-48ca-8534-3f93e4555494",
		//		"9c24b205-32b9-4559-8374-a6fe78db5ca6",
		//		"6838f982-b85d-417f-a454-ec99b90e159b",
		//		"1ee55f17-cc36-44de-9396-2c285d293104",
		//		"9583f93f-49cc-4024-94b0-c24f4f0c85bd",
		//		"8ba75862-2e0d-4e36-a80b-c92076991775",
		//		"b293b678-5d1b-481a-8342-bbd217432513",
		//		"48c3ed1a-3714-45c4-b9ca-ac264e8533ed",
		//		"75687259-963c-43f3-9eab-19124e9f9633",
		//		"50a6aded-416d-4b1b-b0c9-0717075eda65",
		//		"6e593302-42f0-4764-9450-aa9eb939d21b",
		//		"2f41dfdc-370a-4869-9485-5b05870a0fe9",
		//		"31d59523-3a4c-4e76-96d8-81c885324f3b",
		//		"1d509d27-774e-4a13-b71a-6984a5777aec",
		//		"938675c3-219b-4d99-8507-598d36de8e37",
		//		"b99158d6-3eab-48a2-9696-452a06abba2a",
		//		"fa06fd5f-d2ab-4999-a3fd-b9997b06b039",
		//		"d34d5d71-ca92-4263-ad3c-34ff2769dc36",
		//		"2db18cf2-09ab-4c21-870d-32de5caa2a0b",
		//		"beb2ca24-e9a7-45ab-a26c-f59d71738cfa",
		//		"79c1fb88-60b4-434c-8fe5-7fb6ca142adb",
		//		"dd79ac1b-c310-41e2-bdd5-303d5747a2da",
		//		"b7f95ad4-acb7-4e66-aec8-6f6fdbbca588",
		//		"ff1b5b96-8004-49fa-a36b-059e885d571d",
		//		"1c1506f6-6fe8-4ea7-a72a-4afef69dff3a",
		//		"0968b87c-d80b-49cf-8eec-53e65afbacd3",
		//		"a19298f3-3d02-432d-bd02-0d6b4bf0281f",
		//		"ecb44015-153d-42bd-84bd-6fc6df5190a5",
		//		"1d2a81da-5393-4fc0-a016-23be192f624a",
		//		"e331c8fe-fcf0-40ba-9f7b-7c6c95aaf54d",
		//		"f2cb8e81-8ade-4141-bce8-f2a763762af8",
		//		"6b7f15ba-3c9b-4165-9755-9543b52d4bbc",
		//		"d1408065-d31e-43a7-902c-0b66f2327c16",
		//		"d34de710-2b4d-4f16-9cd6-42d48974e2f6",
		//		"fc6cb619-0c67-432c-a43b-5144aa499050",
		//		"77f97cb9-b37a-42a5-bd8d-866c983f9508",
		//		"aa8147e3-6d8e-43d9-8527-47579cf3aa8d",
		"e9cbd134-9066-4a05-a366-b8573ebb53d0", // this works ?
		//		"9294441a-9b8c-47d0-ba68-019557e1cc8b",
		//		"412ba028-9507-4751-ae46-a5586b9dc412",
		//		"5c456a67-c98f-4f0f-8cbc-b5ae85b1a224",
		//		"80930014-dff1-4b06-964d-77f98286e52d",
		//		"572f5960-7461-41bb-b587-3af9cae1c918",
		//		"5f6ba7a9-9655-43d3-8202-b9d22eb98ca3",
		//		"a71f566b-d355-4adc-b366-aa27f81fe380",
		//		"d4f51a35-a594-4c8a-9c56-baac35448ac8",
		//		"2f98ec99-b575-438b-99e8-711291a661e9",
		//		"10e87134-e291-4c0c-9df7-7aeaf60c3d72",
		//		"acbe3399-8060-462b-b61d-77e8a2e3ae32",
		//		"3b130126-6cc4-403b-9097-f40c6c7fb7e1",
		//		"4591067a-480c-4eb2-a240-f88fa1eecf47",
		//		"0e50c031-5a78-40dc-8dbc-ed9ceccecd26",
		//		"ba50b154-862e-4747-993b-47b2a53179f8",
		//		"62cb6df4-f026-4794-92ee-639da82ceb22",
		//		"e10960c3-afef-4e45-86bd-bab87a86e47d",
		//		"3a32c24f-f6ee-420b-a1a0-9a2207afd938",
		//		"dd6bc12d-66eb-4e9c-a055-20a29898bc4e",
		//		"6fc5e1e5-68ee-4157-85f6-f48b8852596c",
		//		"f6f916cf-2fcf-4f37-8fbe-394e3fdb3de7",
		//		"4320bd94-93f6-47d0-87a9-1754d2fe94e8",
		//		"4d57e5a3-5763-4748-84d8-3d8404988134",
		//		"5060ea96-fcc9-4d16-9a20-8544f02f05ca",
		//		"f78978e5-2cec-4593-9a45-94f745083942",
		//		"0a1dc710-349c-4c86-ad83-dd4b417063ac",
		//		"39f8e053-dbef-449e-aa57-07c1bbe1e512",
		//		"6fb1d36d-58a3-483f-92b7-05b9d25734b9",
		//		"c843b138-6472-4608-b001-e9a2e488d439",
		//		"09140900-a1a7-4ccf-a07e-ec7808c6dc8e",
		//		"ebf4e79d-06da-4d87-88aa-07d76014c3c7",
		//		"d0d3456f-1bbb-463c-adf1-e9bb8b919c42",
		//		"d1835518-984b-4d4c-9b38-88e338d9417b",
		"85578044-27a2-4126-8e98-176a92a6a54c", // this works ?
		"b6fb647b-3dd8-4b2e-98bb-533577580938", // this works ?
		//		"fb1871d4-521a-4ae3-9e35-a8fd760b3ca6",
		//		"6b1202e9-4b3b-48e9-a5bd-9cfc33f12bcb",
		//		"7c74009f-2116-4406-a225-bad774664454",
		"92b35c5d-ea94-4622-be87-c25d672faf3e", // this works ?
		//		"d14a8c61-af7c-48f4-ad65-efd653371513",
		//		"cab7e6cd-473d-417a-8d14-2c37b98b8b63",
		//		"03bf5ddf-779b-41be-86a9-23956e1c2ed4",
		//		"7164500c-842a-42aa-a0a0-cf91a5fe7bac",
		//		"4542dd6f-692b-483e-9a98-71245125c456",
		//		"c3295204-837e-4a97-afb3-6135a3ff89f7",
		//		"8b9e17c2-7757-40a8-b50e-3a7053418eac",
		//		"fdee5829-a0ba-4b82-947a-698b2bd2772d",
		//		"73fa9805-eb30-4033-b2a9-c86b103d7419",
		//		"13ac89d9-4e67-4650-a193-bdd6869dd837",
		//		"e6ffd42d-a6c2-4946-956f-a87722e4eece",
		//		"ea497b88-57e3-4f68-984c-a39199364d77",
		//		"d9b97737-dd07-4e4c-ad89-03badc51a045",
		//		"5fffd6eb-d50a-42c6-af2d-57a0243515d2",
		//		"5ea1aae1-bf28-4988-98a7-bf790e3c8960",
		//		"5931794e-1d2f-427f-a699-8cda326aab3d",
		//		"a31cd35d-fcb4-41c8-9160-2c1a2a45e311",
		//		"fa215e00-5724-4a53-b4a4-7604f716b0cd",
		"29cda6a5-0bf8-4121-8854-2cde691c31da", // this works ?
		//		"e741ffd9-9235-4333-8f55-9ed3137af71b",
		//		"a319c4da-3d90-4742-abfc-5710cf6a5321",
	}

	for i := 0; i < 50; i++ {
		//		var count int
		// launch a load (possibly: 269) of filter requests to do performance timings.
		for _, id := range filterIds {
			sendFilter(ctx, kafkaProducer, id)
			//			count++
			//			if count >= 1 {
			//				break
			//			}
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		filterID := scanner.Text()

		sendFilter(ctx, kafkaProducer, filterID)
		time.Sleep(1 * time.Second)
	}
}

func sendFilter(ctx context.Context, kafkaProducer *kafka.Producer, filterID string) {
	log.Event(ctx, "sending filter output event", log.INFO, log.Data{"filter_ouput_id": filterID})

	event := event.FilterSubmitted{
		FilterID: filterID,
	}

	bytes, err := schema.FilterSubmittedEvent.Marshal(event)
	if err != nil {
		log.Event(ctx, "filter submitted event error", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
	kafkaProducer.Initialise(ctx)
	kafkaProducer.Channels().Output <- bytes
}
