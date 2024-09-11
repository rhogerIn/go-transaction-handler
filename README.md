# **Technical Challenge: Credit Card Transaction Authorization**

### **Introduction**

Authorizing a credit card transaction is a crucial part of any banking system. This challenge aims to evaluate different strategies for implementing this key functionality.

---

## **Transaction**

A simplified version of a credit card transaction payload looks like this:

```json
{
  "account": "123",
  "totalAmount": 100.00,
  "mcc": "5811",
  "merchant": "JONH'S RESTAURANT               NYC USA"
}
```
The MCC contains the classification of the merchant. Based on its value, you should decide which balance will be used (for the entire transaction amount). For simplicity, let’s use the following rules:

If the mcc is "5411" or "5412", the FOOD balance should be used.
If the mcc is "5811" or "5812", the MEAL balance should be used.
For any other mcc values, the CASH balance should be used.

