# product

goを勉強する場所

## コマンド

```bash
go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/modifier,sql/execquery,sql/lock,sql/upsert ./ent/schema # entのコード生成
```

## シーダー

```bash
docker compose exec db mysql -u root -p -D product
use product;
insert into products (title, stock, sale_status) values ('book', 100, '0'), ('orange', 200, '1');
curl -X POST -H "Content-Type: application/json" -d '{"product_id" : 1 , "amount" : 1}' localhost:8080/products
curl localhost:8080/products | jq .
{
                "Id": 1,
                "Title": "book",
                "Stock": 99,
                "SaleStatus": 0
}% 
```

## 参考

### プロジェクト構成

- [Go言語で基本的なCRUD操作を行うREST APIを作成](https://dev.classmethod.jp/articles/go-sample-rest-api/)

### トランザクション

- [「トランザクション張っておけば大丈夫」と思ってませんか？ バグの温床になる、よくある実装パターン](https://zenn.dev/tockn/articles/4268398c8ec9a9)
- [MySQLで発生し得る思わぬデッドロックと対応方法](https://zenn.dev/shuntagami/articles/ea44a20911b817)

### テストコード

- [Goのテーブル駆動テストをわかりやすく書きたい](https://zenn.dev/kimuson13/articles/go_table_driven_test)
