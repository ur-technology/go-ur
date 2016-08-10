function unlockAccounts(accounts) {
  accounts.map(function(a) { return personal.unlockAccount(a, "password"); });
}

function describeAccounts(accounts, name) {
  console.log("There are " + accounts.length + " " + name + " accounts, with UR balances " + accounts.map(function(a) { return web3.fromWei(eth.getBalance(a), 'ether'); }));
}

var privilegedAccts = eth.accounts.slice(0,4);
unlockAccounts(privilegedAccts);
describeAccounts(privilegedAccts, "privileged");

// Create 7 mining accounts and 7 recipient accounts.
function createAccounts(numAccounts) {
  var newAccounts = [];
  for (var i = 0; i < numAccounts; i++) {
    newAccount = personal.newAccount("password");
    newAccounts.push(newAccount);
  }
  return newAccounts;
}

var miners = createAccounts(4);
unlockAccounts(miners);
describeAccounts(miners, "miner");

var recipients = createAccounts(4);
describeAccounts(recipients, "recipient");

function mineBlocks(numBlocksToSleep) {
  for (var m=0; m < miners.length; m++) {
    miner.setEtherbase(miners[m]);
    miner.start(7);
    console.log("miner " + m + ": started mining");

    var bal;
    while(true) {
      admin.sleepBlocks(numBlocksToSleep);
      bal = eth.getBalance(miners[m]);
      if (bal > 0) {
        break;
      }
    }

    miner.stop();
    console.log("miner " + m + ": stopped mining because balance is now " + bal);
  }
}

mineBlocks(4);

console.log("mineBlocks finished");

// for reach privileged account, send each recipient a privileged transaction with
// amount of 1 wei (this should trigger a reward of 0.001 UR)
for (var r=0; r < recipients.length; r++) {
  for (var p=0; p < privilegedAccts.length; p++) {
    eth.sendTransaction({from: privilegedAccts[p], to: recipients[r], value: 1, gas: "21000"});
  }
}

// for reach miner account, send each recipient a standard transaction with
// amount of 3 wei
for (var r=0; r < recipients.length; r++) {
  for (var m=0; m < miners.length; m++) {
    eth.sendTransaction({from: miners[m], to: recipients[r], value: web3.toWei(0.003,'ether'), gas: "21000"});
  }
}

// start mining again in order to monitor how many transactions are created per block
console.log("Starting to mine again!")
miner.start(6);

while(true) {
  admin.sleepBlocks(1);
  var b = "";
  var n = eth.blockNumber
  for (var i = 1; i <= n; i++) {
    b = b + ", " + i + ":" + eth.getBlock(i).transactions.length;
  }
  console.log("Transaction count for each block: " + b);

  var pending =  + (eth.pendingTransactions || []).length;
  console.log(pending + " pending transactions remaining");
  if (pending == 0 || n > 100) {
    break;
  }
}

miner.stop();

describeAccounts(privilegedAccts, "privileged");
describeAccounts(miners, "miner");
describeAccounts(recipients, "recipient");
