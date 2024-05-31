CREATE DATABASE bank_v1;
USE bank_v1;

CREATE TABLE userdetails (
    userId SERIAL PRIMARY KEY,
    userName VARCHAR(100),
    userBirthday DATE,
    userGender VARCHAR(10),
    userEmail VARCHAR(100),
    userPassword VARCHAR(100)
);

CREATE TABLE useraddress (
    userId INTEGER,
    userAddressId SERIAL PRIMARY KEY,
    userStreetAddress VARCHAR(100),
    userCity VARCHAR(100),
    userState VARCHAR(100),
    userCountry VARCHAR(100)
);

CREATE TABLE useraccount (
    userId INTEGER,
    userAccountNumber SERIAL PRIMARY KEY,
    userAccountPassword VARCHAR(100),
    userAccountCurrentMoney FLOAT,
    userAccountCreditLimit FLOAT
);
CREATE INDEX idx_userdetails_userId ON userdetails (userId);

CREATE TABLE transactions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    sender_id INT NOT NULL,
    recipient_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    description VARCHAR(255),
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES userdetails(userId),
    FOREIGN KEY (recipient_id) REFERENCES userdetails(userId)
);

CREATE TABLE recipients (
    userId INTEGER,
    recipientId SERIAL PRIMARY KEY,
    recipientName VARCHAR(100),
    recipientAccountNumber INTEGER,
    additionalInfo TEXT
);

CREATE TABLE creditcards (
    cardId SERIAL PRIMARY KEY,
    userId INTEGER,
    cardNumber VARCHAR(16),
    expirationDate DATE,
    cvv VARCHAR(4),
    additionalInfo TEXT
);

ALTER TABLE transactions
ADD FOREIGN KEY (userId) REFERENCES userdetails(userId);

-- Trigger to update sender's balance after a transaction
DELIMITER //

CREATE TRIGGER after_transaction_update_sender_balance
AFTER INSERT ON transactions
FOR EACH ROW
BEGIN
    UPDATE useraccount
    SET userAccountCurrentMoney = userAccountCurrentMoney - NEW.amount
    WHERE userId = NEW.sender_id;
END;
//

DELIMITER ;

-- Trigger to update recipient's balance after a transaction
DELIMITER //

CREATE TRIGGER after_transaction_update_recipient_balance
AFTER INSERT ON transactions
FOR EACH ROW
BEGIN
    UPDATE useraccount
    SET userAccountCurrentMoney = userAccountCurrentMoney + NEW.amount
    WHERE userId = NEW.recipient_id;
END;
//

DELIMITER ;

-- Populate userdetails table with sample data
INSERT INTO userdetails (userName, userBirthday, userGender, userEmail, userPassword)
VALUES 
    ("Joao Augusto", "2005-03-02", "Man", "joao@email.com", "Senha123"),
    ("Thaynara Buratti", "2006-12-01", "Female", "thaynara@email.com", "Senha123");

-- Populate useraccount table with sample data
INSERT INTO useraccount (userId, userAccountPassword, userAccountCurrentMoney, userAccountCreditLimit)
VALUES 
    (1, "1234", 100.0, 1000.0),
    (2, "1234", 100.0, 1000.0);

-- Populate creditcards table with sample data
INSERT INTO creditcards (userId, cardNumber, expirationDate, cvv)
VALUES 
    (1, "0001", "2030-01-01", "0001");


