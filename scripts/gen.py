import csv
import random

from faker import Faker
from faker.providers import bank

ACCOUNTS = [100, 500, 1000, 5000, 10000, 50000, 100000]
BANKS = [
    "CaixaBank",
    "Santander",
    "BBVA",
    "Sabadell",
    "CecaBank",
    "ING",
    "Cooperativo",
    "Bankinter",
    "Cajamar",
    "IberCaja",
    "Kutxa",
    "BancaMarch",
    "Deutsche Bank",
]

with open("data/banks.csv", "w") as csvfile:
    writer = csv.writer(csvfile)
    writer.writerow(["bank"])
    for bank in BANKS:
        writer.writerow([bank])


def generate_account(fake: Faker, fake_bank: Faker) -> list:

    iban = fake_bank.iban()
    bank_name = random.choice(BANKS)
    name = fake.name()
    ammount = random.randrange(100, 10000)

    return [iban, bank_name, name, ammount]


for n_accounts in ACCOUNTS:
    print(f"Generating {n_accounts} accounts")
    with open(f"data/accounts_{n_accounts}.csv", "w") as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow(["iban", "bank", "name", "ammount"])

        fake = Faker(["es_ES"])
        fake_bank = Faker(["es_ES"])
        fake_bank.add_provider(bank)
        for n in range(0, n_accounts):
            writer.writerow(generate_account(fake, fake_bank))
