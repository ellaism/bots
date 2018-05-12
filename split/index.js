const BYTECODE = "0x6060604052341561000f57600080fd5b6040516060806103748339810160405280805191906020018051919060200180519150505b8060018110158015610047575060638111155b151561005257600080fd5b60018054600160a060020a03808716600160a060020a031992831617909255600080549286169290911691909117905560028290555b5b505050505b6102d78061009d6000396000f300606060405236156100465763ffffffff60e060020a60003504166308b27e3e81146100d357806370ba1113146101065780638da5cb5b1461012b578063ed88c68e1461015a575b5b60008060003411156100cd5760025460649034025b6001549190049250348390039150600160a060020a031681156108fc0282604051600060405180830381858888f19350505050151561009a57600080fd5b600054600160a060020a031682156108fc0283604051600060405180830381858888f1935050505015156100cd57600080fd5b5b5b5050005b34156100de57600080fd5b6100f2600160a060020a0360043516610189565b604051901515815260200160405180910390f35b341561011157600080fd5b610119610287565b60405190815260200160405180910390f35b341561013657600080fd5b61013e61028d565b604051600160a060020a03909116815260200160405180910390f35b341561016557600080fd5b61013e61029c565b604051600160a060020a03909116815260200160405180910390f35b60008181600160a060020a0382166370a0823130836040516020015260405160e060020a63ffffffff8416028152600160a060020a039091166004820152602401602060405180830381600087803b15156101e357600080fd5b6102c65a03f115156101f457600080fd5b5050506040518051600154909250600160a060020a03808516925063a9059cbb91168360006040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526024820152604401602060405180830381600087803b151561026357600080fd5b6102c65a03f1151561027457600080fd5b50505060405180519350505b5050919050565b60025481565b600154600160a060020a031681565b600054600160a060020a0316815600a165627a7a723058205ac0d45b075c0c38b2d8ce4d943ab372e06782768fc850ee607d261aa9d2833e0029";
const ABI = [{"constant":false,"inputs":[{"name":"tokenAddress","type":"address"}],"name":"transferAnyERC20Token","outputs":[{"name":"success","type":"bool"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"percent","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"donate","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"inputs":[{"name":"_owner","type":"address"},{"name":"_donate","type":"address"},{"name":"_percent","type":"uint256"}],"payable":false,"type":"constructor"},{"payable":true,"type":"fallback"}];

const Discord = require("discord.js");
const Web3 = require('web3');

const discordClient = new Discord.Client();
const web3 = new Web3(new Web3.providers.HttpProvider(process.env.RPC_HOST));

const splitContract = new web3.eth.Contract(ABI);

discordClient.on("message", async message => {
  if (message.author.bot) return;

  if (message.content.indexOf("!split") === 0) {
    const args = message.content.slice("!split".length).trim().split(/ +/g);

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
      donateAddr = "0x01ff0FFd25B64dE2217744fd7d4dc4aA3cAbceE7";
    } else if (donate == "devpool") {
      donateAddr = "0x65767ec6d4d3d18a200842352485cdc37cbf3a21";
    } else if (donate == "default") {
      donateAddr = "0xEDae451f57B5bfF81d1D9eE64F591Ad6a865a652";
    } else if (donate == "token") {
      donateAddr = "0x4aaad871293c4581edb580e99fb6613b0a3bc488";
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
  } else if (message.content.indexOf("!mining") === 0) {
    const args = message.content.slice("!mining".length).trim().split(/ +/g);

    var address;

    if (args.length == 2 && args[0] == "withdraw") {
      address = args[1];
    } else {
      message.reply(`${message.author.tag} Invalid command.`);
      return;
    }

    if (!web3.utils.isAddress(address)) {
      message.reply(`${message.author.tag} Invalid split contract address.`);
      return;
    }

    try {
      var callSplitContract = new web3.eth.Contract(ABI);

      callSplitContract.options.address = address;
      callSplitContract.methods.transferAnyERC20Token("0x991e7fe4b05f2b3db1d788e705963f5d647b0044")
        .send({
          from: "0x9e2d4a8116c48649ff8b26a67e3f8e4b9ed7cef6",
          gasPrice: 0
        }, function(err, transactionHash) {
          if (err) {
            message.reply(`${message.author.tag} Something wrong happened. Please try again.`);
          } else {
            message.reply(`${message.author.tag} All MINING tokens sent to the owner in transaction ${transactionHash}`);
          }
        });
    } catch(e) {
      console.log(e);
      message.reply(`${message.author.tag} Sorry but we are having some problem withdrawing tokens this time.`);
    }
  }
});

discordClient.login(process.env.DISCORD);
