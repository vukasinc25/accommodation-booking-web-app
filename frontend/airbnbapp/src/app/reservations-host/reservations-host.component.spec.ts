import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ReservationsHostComponent } from './reservations-host.component';

describe('ReservationsHostComponent', () => {
  let component: ReservationsHostComponent;
  let fixture: ComponentFixture<ReservationsHostComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ReservationsHostComponent]
    });
    fixture = TestBed.createComponent(ReservationsHostComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
