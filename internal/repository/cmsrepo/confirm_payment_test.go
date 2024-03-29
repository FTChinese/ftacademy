package cmsrepo

import (
	"context"
	"github.com/FTChinese/ftacademy/internal/mock"
	"github.com/FTChinese/ftacademy/internal/pkg/checkout"
	"github.com/FTChinese/ftacademy/pkg/db"
	"github.com/FTChinese/ftacademy/pkg/faker"
	"github.com/FTChinese/ftacademy/pkg/price"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_listLicenceTxnPrice(t *testing.T) {

	repo := mock.NewRepo()

	schema := mock.NewAdmin().CartBuilder().
		AddNewStandardN(10).
		AddRenewalStandardN(10).
		AddNewPremiumN(5).
		AddRenewalPremiumN(5).
		BuildOrderSchema()

	repo.InsertOrderSchema(schema)

	env := NewEnv(db.MockMySQL(), zaptest.NewLogger(t))

	priceCh := env.listLicenceTxnPrice(schema.OrderRow.ID)

	for pt := range priceCh {
		t.Logf("%v", pt)
	}
}

func TestEnv_buildLicence(t *testing.T) {
	repo := mock.NewRepo()
	adm := mock.NewAdmin()

	granted := adm.StdLicenceBuilder().
		SetPersona(mock.NewPersona()).
		BuildGranted()

	schema := adm.CartBuilder().
		AddNewStandardN(1).
		AddRenewal(granted.ExpLicence).
		BuildOrderSchema()

	repo.CreateGrantedLicence(granted)
	repo.InsertOrderSchema(schema)

	env := NewEnv(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		txnPrice checkout.PriceOfLicenceTxn
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Build new licence",
			args: args{
				txnPrice: checkout.PriceOfLicenceTxn{
					TxnID: schema.Transactions[0].ID,
					Price: price.MockPriceStdYear,
				},
			},
			wantErr: false,
		},
		{
			name: "Build renewal licence",
			args: args{
				txnPrice: checkout.PriceOfLicenceTxn{
					TxnID: schema.Transactions[1].ID,
					Price: price.MockPriceStdYear,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.buildLicence(tt.args.txnPrice)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildLicence() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func getNumber(to int) <-chan int {
	ch := make(chan int)

	go func() {
		defer close(ch)

		for i := 0; i < to; i++ {
			ch <- i
		}
	}()

	return ch
}

func TestSemaphore(t *testing.T) {
	t.Logf("Max workers: %d", maxWorkers)
	intCh := getNumber(1000)
	ctx := context.Background()
	for i := range intCh {
		if err := sem.Acquire(ctx, 1); err != nil {
			t.Error(err)
			break
		}

		go func(n int) {
			t.Logf("Handle %d", n)
			sem.Release(1)
		}(i)
	}

	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		t.Error(err)
		return
	}

	t.Log("Finished")
}
