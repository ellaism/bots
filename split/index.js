const BYTECODE = "0x6060604052341561000f57600080fd5b6040516060806102518339810160405280805191906020018051919060200180519150505b8060018110158015610047575060638111155b151561005257600080fd5b60018054600160a060020a03808716600160a060020a031992831617909255600080549286169290911691909117905560028290555b5b505050505b6101b48061009d6000396000f300606060405236156100545763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166370ba111381146100e15780638da5cb5b14610106578063ed88c68e14610135575b5b60008060003411156100db5760025460649034025b6001549190049250348390039150600160a060020a031681156108fc0282604051600060405180830381858888f1935050505015156100a857600080fd5b600054600160a060020a031682156108fc0283604051600060405180830381858888f1935050505015156100db57600080fd5b5b5b5050005b34156100ec57600080fd5b6100f4610164565b60405190815260200160405180910390f35b341561011157600080fd5b61011961016a565b604051600160a060020a03909116815260200160405180910390f35b341561014057600080fd5b610119610179565b604051600160a060020a03909116815260200160405180910390f35b60025481565b600154600160a060020a031681565b600054600160a060020a0316815600a165627a7a72305820eed8e56da1b6a717c7f16bd998122eaaed458d92c335f02dc0e7ab558a1d7fbe0029";
const ABI = [{"constant":true,"inputs":[],"name":"percent","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"donate","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"inputs":[{"name":"_owner","type":"address"},{"name":"_donate","type":"address"},{"name":"_percent","type":"uint256"}],"payable":false,"type":"constructor"},{"payable":true,"type":"fallback"}];

const PREFIX = "!split";

const Discord = require("discord.js");
const Web3 = require('web3');

const discordClient = new Discord.Client();
const web3 = new Web3(new Web3.providers.HttpProvider(process.env.RPC_HOST));

const splitContract = new web3.eth.Contract(ABI);

discordClient.on("message", async message => {
  if (message.author.bot) return;
  if (message.content.indexOf(PREFIX) !== 0) return;

  const args = message.content.slice(PREFIX.length).trim().split(/ +/g);

  var donate;
  var percent;
  var owner;

  if (args.length == 3) {
    donate = args[0];
    percent = parseInt(args[1]);
    owner = args[2];
  } else if (args.length == 2) {
    donate = "default";
    percent = parseInt(args[0]);
    owner = args[1];
  } else {
    message.reply(`${message.author.tag} Invalid command.`);
    return;
  }

  var donateAddr = "0xe9C2d958E6234c862b4AfBD75b2fd241E9556303";

  if (donate === "dev") {
    donateAddr = "0xe9C2d958E6234c862b4AfBD75b2fd241E9556303";
  } else if (donate == "community") {
    donateAddr = "0xA2C7779077Edc618C926AB5BA7510877C187cd92";
  } else if (donate == "default") {
    donateAddr = "0xCDc7A9C589658fD31fb7ACd3765f02118e4C15Ff";
  } else {
    message.reply(`${message.author.tag} Unknown donation target.`);
    return;
  }

  if (!(percent >= 1 && percent <= 99)) {
    message.reply(`${message.author.tag} Invalid percent.`);
    return;
  }
  if (!web3.utils.isAddress(donateAddr)) {
    message.reply(`${message.author.tag} Invalid donation address.`);
    return;
  }
  if (!web3.utils.isAddress(owner)) {
    message.reply(`${message.author.tag} Invalid owner address.`);
    return;
  }

  await message.reply(`${message.author.tag} We're creating the contract for you. Please wait...`);
  try {
    console.log(`Creating address with owner ${owner}, donate ${donateAddr} and percent ${percent}.`);
    const newContractInstance = await splitContract.deploy({
      data: BYTECODE,
      arguments: [owner, donateAddr, percent],
    }).send({
      from: "0x9e2d4a8116c48649ff8b26a67e3f8e4b9ed7cef6",
      gasPrice: 0,
    });
    message.reply(`${message.author.tag} Your address is ready at ${newContractInstance.options.address}. This donates ${percent}% to ${donate}.`);
  } catch(e) {
    console.log(e);
    message.reply(`${message.author.tag} Sorry but we are having some problem getting your address this time.`);
  }
});

discordClient.login(process.env.DISCORD);
